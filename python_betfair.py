import requests
import json
import datetime
import sys
 
#url="https://api.betfair.com/exchange/betting/json-rpc/v1"
endpoint = "https://api.betfair.com/exchange/betting/rest/v1.0/"
header = { 'X-Application' : 'Va2gI0Bow1LUDFpd', 'X-Authentication' : 'DLwu3rsqQR6din6H1/MGKk+QnrLq8w9tSyilyQczXcI=' ,'content-type' : 'application/json' }
 
#jsonrpc_req='{"jsonrpc": "2.0", "method": "SportsAPING/v1.0/listEventTypes", "params": {"filter":{ }}, "id": 1}'
json_req='{"filter":{ }}'
 
url = endpoint + "listEventTypes/"


#response = requests.post(url, data=jsonrpc_req, headers=header)
response = requests.post(url, data=json_req, headers=header)
 
print(json.dumps(json.loads(response.text), indent=3))




def getMarketCatalouge(eventTypeID):
    if(eventTypeID is not None):
        print('Calling listMarketCatalouge Operation to get MarketID and selectionId')

        endpoint = "https://api.betfair.com/exchange/betting/rest/v1.0/listMarketCatalogue/"
        header = { 'X-Application' : 'Va2gI0Bow1LUDFpd', 'X-Authentication' : 'DLwu3rsqQR6din6H1/MGKk+QnrLq8w9tSyilyQczXcI=' ,'content-type' : 'application/json' }
 
        now = datetime.datetime.now().strftime('%Y-%m-%dT%H:%M:%SZ')
        print("time is " +str(now))
        market_catalouge_req = '{"filter":{"eventTypeIds":["' + eventTypeID + '"],"marketTypeCodes":["WIN"],"marketCountries":["GB"],"marketStartTime":{"from":"' + now + '"}},"sort":"FIRST_TO_START","maxResults":"1","marketProjection":["RUNNER_METADATA"]}'

        response = requests.post(endpoint, data=market_catalouge_req, headers=header)
  
        print(json.dumps(json.loads(response.text), indent=3))
        market_catalouge_loads = json.loads(response.text)
        return market_catalouge_loads
 
 
def getMarketId(marketCatalougeResult):
    if(marketCatalougeResult is not None):
        for market in marketCatalougeResult:
            return market['marketId']
 
 
def getSelectionId(marketCatalougeResult):
    if(marketCatalougeResult is not None):
        for market in marketCatalougeResult:
            return market['runners'][0]['selectionId']





marketCatalougeResult = getMarketCatalouge('7')
marketid = getMarketId(marketCatalougeResult)
print(marketid)
runnerId = getSelectionId(marketCatalougeResult)
print(runnerId)