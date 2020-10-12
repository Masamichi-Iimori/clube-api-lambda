package main

import (
	"errors"
	"encoding/json"

	// "io/ioutil"
	// "net/http"
	"fmt"
	"net/url"
	"os"

	"github.com/ChimeraCoder/anaconda"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("Non 200 Response found")
)

type User struct {
	Id int64  `json:"id"`
	Name string  `json:"name"`
}

type Tweet struct {
	Id int64  `json:"id"`
	FullText   string `json:"full_text"`
	User User `json:"user"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println(os.Getenv("CONSUMER_KEY"))
	
	anaconda.SetConsumerKey(os.Getenv("CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("CONSUMER_SECRET"))

	api := anaconda.NewTwitterApi(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_TOKEN_SECRET"))

	v := url.Values{}
	v.Set("tweet_mode", "extended")

	searchResult, _ := api.GetSearch("#プロクラブ", v)

	tweets := make([]Tweet, 0)

	for _, tweet := range searchResult.Statuses {
		// リツイートされたものはfull_textでも'RT <ユーザ名>'が入るので省略されてしまうので、その判定
    if tweet.RetweetedStatus == nil {
			tweets = append(tweets, Tweet{tweet.Id, tweet.FullText, User{tweet.User.Id, tweet.User.Name}})
		}else{
			tweets = append(tweets, Tweet{tweet.Id, tweet.RetweetedStatus.FullText, User{tweet.User.Id, tweet.User.Name}})
		}
	}
	jsonBytes, _ := json.Marshal(tweets)


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
