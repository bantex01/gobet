package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"time"

	"github.com/bantex01/gobetUtils"
)

/* EventStruct response
{
    "eventType": {
        "id": "468328",
        "name": "Handball"
        },
    "marketCount": 11
}
*/

type EventStruct struct {
	EventType struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"eventType"`
	MarketCount int `json:"marketCount"`
}

type UserCheckEventStruct struct {
	EventType string
	Id        string
	TimeRange string
	Cycle     int
	Test      string
}

/* MarketCatalogue response
{
	"marketId": "1.200104587",
	"marketName": "7f Hcap",
	"totalMatched": 59854.69
}
*/

type MarketCatalogueResponse struct {
	MarketId     string  `json:"marketId"`
	MarketName   string  `json:"marketName"`
	TotalMatched float32 `json:"totalMatched"`
	Runners      []struct {
		SelectionID int    `json:"selectionId"`
		RunnerName  string `json:"runnerName"`
		Meta        struct {
			RunnerId string `json:"runnerId"`
		} `json:"metadata"`
	} `json:"runners"`
}

var runnerMap = make(map[int]string)

func main() {

	//done := make(chan bool)
	//go gobetUtils.MarketReturn(done)
	//noOfWorkers := 5
	//gobetUtils.CreateMarketWorkerPool(noOfWorkers)

	body := gobetUtils.SubmitAPIRequest("config.yaml", "EXCHANGE", "POST", "listEventTypes/", `{"filter":{}}`)

	sb := string(body)
	log.Println(sb)

	var output []EventStruct

	//merr := json.Unmarshal([]byte(body), &output)
	err := json.Unmarshal(body, &output)
	if err != nil {
		print(err)
	}

	// Let's create a map of event types so we can check on them later
	var eventTypeMap = make(map[string]string)
	for _, value := range output {
		//fmt.Printf("ID:%v - Name:%v - Market Count:%d\n", value.EventType.Id, value.EventType.Name, value.MarketCount)
		eventTypeMap[value.EventType.Name] = value.EventType.Id
	}

	// We have event types, let's find horse racing

	/*
		var eventId string
		re := regexp.MustCompile(`Horse Racing`)
		for _, value := range output {
			fmt.Printf("ID:%v - Name:%v - Market Count:%d\n", value.EventType.Id, value.EventType.Name, value.MarketCount)
			//string := string(value.EventType.Name)
			//match, _ := regexp.MatchString("Horse Racing", value.EventType.Name)
			//if match {

			if re.MatchString(value.EventType.Name) {
				fmt.Printf("Horse Racing found, event id is %s\n", value.EventType.Id)
				eventId = value.EventType.Id
			}
		}

	*/

	// Now let's gather the config and place that in a slice of structs for processing

	var UserCheckEvents = make([]UserCheckEventStruct, 0)
	count := 0
	for _, event := range gobetUtils.Config.EventTypes {

		fmt.Printf("event is %v, lets check for the ID\n", event)
		_, isPresent := eventTypeMap[event]
		if isPresent {
			fmt.Printf("%v has been found, id is %v\n", event, eventTypeMap[event])
			UserCheckEvents = append(UserCheckEvents, UserCheckEventStruct{EventType: event, Id: eventTypeMap[event], TimeRange: gobetUtils.Config.TimeRange[count], Cycle: gobetUtils.Config.Cycle[count], Test: gobetUtils.Config.Test[count]})
		}
		count++
	}

	fmt.Printf("user check event slice is %v\n", UserCheckEvents)

	// Now we have our Event ID, we need to find all events of that type (and keep updating it every 60 seconds) in the timescale specified by the config?

	var marketCatalogueOutput []MarketCatalogueResponse
	var trackedMarkets = make(map[string]string)
	//markets := make([]string, 0)

	for range time.Tick(time.Second * 60) {

		for _, userEvent := range UserCheckEvents {
			fmt.Printf("EventType is %v\n", userEvent.EventType)

			tn := time.Now()
			atime := tn.Format(time.RFC3339)

			fmt.Println("Getting market data to ascertain marketID to then track prices...")

			//go gobetUtils.AllocateTrackMarketJob(userEvent.Id)

			filter := `{"filter":{"eventTypeIds":["` + userEvent.Id + `"],"marketTypeCodes":["WIN"],"marketStartTime":{"from":"` + atime + `"}},"sort":"FIRST_TO_START","maxResults":"1","marketProjection":["RUNNER_METADATA"]}`
			//filter := `{"filter":{"eventTypeIds":["` + userEvent.Id + `"],"marketTypeCodes":["MATCH_ODDS"],"marketStartTime":{"from":"` + atime + `"}},"sort":"FIRST_TO_START","maxResults":"1","marketProjection":["RUNNER_METADATA"]}`
			//filter := `{"filter":{"eventTypeIds":["` + userEvent.Id + `"],"marketCountries":["GB"],"marketStartTime":{"from":"` + atime + `"}},"sort":"FIRST_TO_START","maxResults":"10"}`
			//filter := `{"filter":{"eventTypeIds":["` + userEvent.Id + `"],"marketCountries":["GB"],"marketStartTime":{"from":"` + atime + `"}},"sort":"FIRST_TO_START","maxResults":"1","marketProjection":["RUNNER_METADATA"]}`
			//body = bantexConfig.SubmitAPIRequest("config.yaml", "EXCHANGE", "POST", "listMarketCatalogue/", `{"filter":{"eventTypeIds":["7"],"marketCountries":["GB"],"marketStartTime":{"from":"2022-06-12T08:57:34+01:00"}},"sort":"FIRST_TO_START","maxResults":"1","marketProjection":["RUNNER_METADATA"]}`)
			body = gobetUtils.SubmitAPIRequest("config.yaml", "EXCHANGE", "POST", "listMarketCatalogue/", filter)
			sb := string(body)
			log.Println(sb)

			err := json.Unmarshal(body, &marketCatalogueOutput)
			if err != nil {
				fmt.Println(err)
			}

		}

		// We need to keep a track of the market IDs we're looking at so we don't fire off any unnecssary routines to watch them.
		// For each market we aren't tracking we will fire off a new routine

		for _, value := range marketCatalogueOutput {
			fmt.Printf("ID:%v - Name:%v - Matched:%f\n", value.MarketId, value.MarketName, value.TotalMatched)
			fmt.Printf("Looping runners...\n")
			for _, runners := range value.Runners {
				fmt.Printf("Runner: %v - SelectionID: %d - Runner ID: %v\n", runners.RunnerName, runners.SelectionID, runners.Meta.RunnerId)
				//fmt.Printf("converting selection id string to int\n")
				i, err := strconv.Atoi(runners.Meta.RunnerId)
				if err != nil {
					// handle error
					fmt.Println(err)
				}
				runnerMap[i] = runners.RunnerName
				fmt.Printf("added runner: %v to runner map\n", runnerMap[i])
			}
			// let's keep a track of the marketIDs we're following

			//_, isPresent := trackedMarkets[value.MarketId]
			_, found := trackedMarkets[value.MarketId]
			if !found {
				go gobetUtils.TrackMarket(value.MarketId, value.MarketName, runnerMap)
				trackedMarkets[value.MarketId] = "TRACKED"
				//fmt.Println("Already tracking market " + value.MarketId)
			} //else {

			//markets = append(markets, value.MarketId)
			//go gobetUtils.TrackMarket(value.MarketId)
			//trackedMarkets[value.MarketId] = "TRACKED"

			//go gobetUtils.AllocateTrackMarketJob(value.MarketId)
			//go gobetUtils.TrackMarket(value.MarketId)
			//trackedMarkets[value.MarketId] = "TRACKED"
			//go gobetUtils.MarketReturn(done)
			/*
				// let's get a routine going to check on market
				fmt.Println("gathering odds for market...")
				filter := `{"marketIds":["` + value.MarketId + `"],"priceProjection":{"priceData":["EX_BEST_OFFERS"]}, "id": 1}`
				body = gobetUtils.SubmitAPIRequest("config.yaml", "EXCHANGE", "POST", "listMarketBook/", filter)
				sb = string(body)
				log.Println(sb)
				trackedMarkets[value.MarketId] = "TRACKED"
			*/

			//}

		}

		//for _, marketToCheck := range markets {
		//	fmt.Printf("Market to send to job queue is %v\n", marketToCheck)
		//}

		/*
			done := make(chan bool)
			go gobetUtils.MarketReturn(done)
			noOfWorkers := 5
			gobetUtils.CreateMarketWorkerPool(noOfWorkers)

			//<-done
		*/

		fmt.Printf("Tracked markets are: %v\n", trackedMarkets)
	}

	// testing of asking for prices

	/*
		filter := `{"marketIds":["1.200104613"],"priceProjection":{"priceData":["EX_BEST_OFFERS"]}, "id": 1}`
		//body = bantexConfig.SubmitAPIRequest("config.yaml", "EXCHANGE", "POST", "listMarketCatalogue/", `{"filter":{"eventTypeIds":["7"],"marketCountries":["GB"],"marketStartTime":{"from":"2022-06-12T08:57:34+01:00"}},"sort":"FIRST_TO_START","maxResults":"1","marketProjection":["RUNNER_METADATA"]}`)
		body = gobetUtils.SubmitAPIRequest("config.yaml", "EXCHANGE", "POST", "listMarketBook/", filter)

		sb = string(body)
		log.Println(sb)
	*/

}
