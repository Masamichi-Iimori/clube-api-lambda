package main

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"

	"local.packages/tweet"
)

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("Non 200 Response found")
)

// パラメータで時間指定があればその時間だけ遡れるといい？
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	region := "ap-northeast-1"
	tableName := "proclub_tweets"

	// 認証
	creds := credentials.NewStaticCredentials(os.Getenv("AWS_ACCEESS_KEY"), os.Getenv("AWS_SECRET_ACCEESS_KEY"), "") //第３引数はtoken

	sess, _ := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(region)},
	)

	db := dynamo.New(sess)
	table := db.Table(tableName)

	// 2時間前までのツイート一覧を取得する
	const pastTime = 2

	var tweets tweet.Tweets

	filterTime := time.Now().Add(time.Duration(1) * time.Hour * -pastTime).Unix()

	err := table.Scan().Filter("'tweeted_at' >= ?", filterTime).All(&tweets)

	if err != nil {
		panic(err)
	}

	sort.Sort(tweets)

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
