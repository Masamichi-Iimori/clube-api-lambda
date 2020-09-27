package main

import (
	"errors"
	"encoding/json"

	// "io/ioutil"
	// "net/http"
	// "fmt"
	// "net/url"
	// "os"

	"github.com/ChimeraCoder/anaconda"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)


var consumerKey = "EULx83TrE6sXOl9XbkyIlaRDy"
var consumerSecret = "MwC4UGc8Fa2m9ulXPvaG0xw5UIycHqoubUESNFwKSmnOVUzvpT"
var accessToken = "835079557665284096-31DMiJd9m2DMb7cNrIG027KJTFOT6E3"
var accessTokenSecret = "ZYEf6aq7aMiS9Of2Yf5vSYvnSqrkrhjZwckzw18B0ZNGi"

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("Non 200 Response found")
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	api := anaconda.NewTwitterApi(accessToken, accessTokenSecret)
	searchResult, _ := api.GetSearch("#プロクラブ", nil)

	jsonBytes, _ := json.Marshal(searchResult.Statuses)


	return events.APIGatewayProxyResponse{
			Headers: map[string]string{
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Headers": "origin,Accept,Authorization,Content-Type",
        "Content-Type":                 "application/json",
			},
			Body:       string(jsonBytes),
			StatusCode: 200,
	}, nil
}

// func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
// 	personID := request.PathParameters["personID"]
// 	personName := request.QueryStringParameters["personName"]
// 	old := 25

// 	person := PersonResponse{
// 			PersonID:   personID,
// 			PersonName: personName,
// 			Old:        old,
// 	}
// 	jsonBytes, _ := json.Marshal(person)

// 	return events.APIGatewayProxyResponse{
// 			Body:       string(jsonBytes),
// 			StatusCode: 200,
// 	}, nil
// }


func main() {
	lambda.Start(handler)
}
