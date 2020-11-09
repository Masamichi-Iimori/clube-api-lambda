package main

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strings"
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

// スライスに指定した要素が含まれているかチェックする関数
func contains(target interface{}, list interface{}) bool {

	switch list.(type) {
	default:
		return false
	case []int:
		revert := list.([]int)
		for _, r := range revert {
			if target == r {
				return true
			}
		}
		return false

	case []uint64:
		revert := list.([]uint64)
		for _, r := range revert {
			if target == r {
				return true
			}
		}
		return false

	case []string:
		revert := list.([]string)
		for _, r := range revert {
			if target == r {
				return true
			}
		}
		return false
	}
}

// 指定したポジションが含まれている、もしくはポジション別の募集が無いツイートだけをfilteredTweetsに入れる
func filteringPositions(filterPositions []string, tweets tweet.Tweets) (result tweet.Tweets) {
	for _, tweet := range tweets {
		isMatch := false
		for _, position := range tweet.Position {
			if contains(position, filterPositions) {
				isMatch = true
			}
		}
		if isMatch || len(tweet.Position) == 0 {
			result = append(result, tweet)
		}
	}

	return result
}

func filteringWords(filterWords []string, tweets tweet.Tweets) (result tweet.Tweets) {
	for _, tweet := range tweets {
		for _, word := range filterWords {
			if strings.Contains(tweet.FullText, word) {
				result = append(result, tweet)
			}
		}
	}
	if result == nil {
		result = tweet.Tweets{}
	}

	return result
}

// パラメータで時間指定があればその時間だけ遡れるといい？
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	query := request.QueryStringParameters
	filterPositions := []string{}
	filterWords := []string{}
	if len(query["positions"]) != 0 {
		filterPositions = strings.Split(query["positions"], ",")
	}
	if len(query["words"]) != 0 {
		filterWords = strings.Split(query["words"], ",")
	}

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

	// 指定した時間までさかのぼってツイート一覧を取得する
	const pastTime = 2

	var tweets tweet.Tweets

	filterTime := time.Now().Add(time.Duration(1) * time.Hour * -pastTime).Unix()

	err := table.Scan().Filter("'tweeted_at' >= ?", filterTime).All(&tweets)

	if err != nil {
		panic(err)
	}

	//filteredTweets := []Tweet{}

	if len(filterPositions) != 0 {
		tweets = filteringPositions(filterPositions, tweets)
	}

	if len(filterWords) != 0 {
		tweets = filteringWords(filterWords, tweets)
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
