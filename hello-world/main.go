package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	// "io/ioutil"
	// "net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("Non 200 Response found")
)

// User つぶやいたユーザ情報
type User struct {
	ID   int64  `dynamo:"id" json:"id"`
	Name string `dynamo:"name" json:"name"`
}

// Tweet 参加を募集するツイート
type Tweet struct {
	ID        int64  `dynamo:"tweet_id" json:"tweet_id"`
	FullText  string `dynamo:"full_text" json:"full_text"`
	TweetedAt int64  `dynamo:"tweeted_at" json:"tweeted_at"`
	User      User   `dynamo:"user" json:"user"`
}

// パラメータで時間指定があればその時間だけ遡れるといい？
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	region := "ap-northeast-1"
	tableName := "proclub_tweets"
	//sess := session.Must(session.NewSession())
	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	table := db.Table(tableName)

	tweets := []Tweet{}
	// １時間前までのツイート一覧を取得する
	filterTime := time.Now().Add(time.Duration(1) * time.Hour * -1).Unix()

	err := table.Scan().Filter("'tweeted_at' >= ?", filterTime).All(&tweets)
	//err := table.Scan().All(&tweets)

	if err != nil {
		panic(err)
	}

	for _, tweet := range tweets {
		log.Println(tweet.FullText)
	}
	jsonBytes, _ := json.Marshal(tweets)

	return events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "origin,Accept,Authorization,Content-Type",
			"Content-Type":                 "application/json",
		},
		Body:       string(jsonBytes),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
