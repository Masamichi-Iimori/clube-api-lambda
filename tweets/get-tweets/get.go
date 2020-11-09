package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	// "io/ioutil"
	// "net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
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
	ID         int64  `dynamo:"id" json:"id"`
	Name       string `dynamo:"name" json:"name"`
	ScreenName string `dynamo:"screen_name" json:"screen_name"`
}

// Tweet 参加を募集するツイート
type Tweet struct {
	ID        int64    `dynamo:"tweet_id" json:"tweet_id"`
	FullText  string   `dynamo:"full_text" json:"full_text"`
	TweetedAt int64    `dynamo:"tweeted_at" json:"tweeted_at"`
	User      User     `dynamo:"user" json:"user"`
	Position  []string `dynamo:"position" json:"positions"`
	IsClub    bool     `dynamo:"is_club" json:"is_club"`
}

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

	tweets := []Tweet{}

	filterTime := time.Now().Add(time.Duration(1) * time.Hour * -pastTime).Unix()

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
