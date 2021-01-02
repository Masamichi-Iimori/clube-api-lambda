package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/garyburd/go-oauth/oauth"
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

// Request APIGatewayからのリクエストを受け取るための構造体
type Request struct {
	OauthToken    string `json:"oauth_token"`
	OauthVerifier string `json:"oauth_verifier"`
	SessionID     string `json:"session_id"`
}

// Token TwitterAPIから取得した一時Tokenを保存するための構造体
type Token struct {
	ID           string `dynamo:"id"`
	OauthToken   string `dynamo:"oauth_token"`
	SecretToken  string `dynamo:"secret_token"`
	RegisterDate string `dynamo:"register_date"`
}

// Session TwitterAPIから取得したアクセストークンを保存するための構造体
type Session struct {
	ID           string `dynamo:"id" json:"id"`
	AccessToken  string `dynamo:"access_token" json:"acceess_token"`
	SecretToken  string `dynamo:"secret_token" json:"secret_token"`
	RegisterDate string `dynamo:"register_date" json:"register_date"`
	ScreenName   string `dynamo:"screen_name" json:"screen_name"`
	UserID       string `dynamo:"user_id" json:"user_id"`
}

// Account TwitterAPIから取得したユーザー情報から、screen_nameを取り出すための構造体
type Account struct {
	ID                   string `dynamo:"id_str" json:"id_str"`
	ScreenName           string `dynamo:"screen_name" json:"screen_name"`
	ProfileImageURL      string `dynamo:"profile_image_url" json:"profile_image_url"`
	ProfileImageURLHttps string `dynamo:"profile_image_url_https" json:"profile_image_url_https"`
}

// Response APIGatewayへのレスポンスを定義するための構造体
type Response struct {
	Location string `json:"location"`
	Cookie   string `json:"cookie"`
}

func createSessionID(screenName string) string {
	str := screenName + time.Now().Format("2006-01-02 15:04:05")
	return base64.URLEncoding.EncodeToString([]byte(str))
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	query := request.QueryStringParameters
	fmt.Println("---------------------")
	fmt.Println(query)
	fmt.Println("---------------------")

	//OAuthの設定
	oauthClient := &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  os.Getenv("CONSUMER_KEY"),
			Secret: os.Getenv("CONSUMER_SECRET"),
		},
		TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authenticate",
		TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
	}

	region := "ap-northeast-1"

	// 認証
	creds := credentials.NewStaticCredentials(os.Getenv("AWS_ACCEESS_KEY"), os.Getenv("AWS_SECRET_ACCEESS_KEY"), "") //第３引数はtoken

	sess, _ := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(region)},
	)

	db := dynamo.New(sess)
	tokenTable := db.Table("token")
	sessionID := query["session_id"]

	// DBからOAuthトークンの取得
	var token []Token
	err := tokenTable.Get("id", sessionID).All(&token)
	if err != nil {
		fmt.Println("err")
		panic(err.Error())
	}

	// OAuthトークンを決められた構造体にする
	tempCredentials := &oauth.Credentials{
		Token:  token[0].OauthToken,
		Secret: token[0].SecretToken,
	}

	// Twitterから返されたOAuthトークンと、あらかじめlogin.goで入れておいたセッション上のものと一致するかをチェック
	if tempCredentials.Token != query["oauth_token"] {
		panic(tempCredentials.Token + "_" + query["oauth_token"])
	}

	//アクセストークンの取得
	tokenCard, _, err := oauthClient.RequestToken(nil, tempCredentials, query["oauth_verifier"])
	if err != nil {
		log.Fatal("RequestToken:", err)
		panic(err.Error())
	}

	// 時間取得時のフォーマット指定
	format := "2006-01-02 15:04:05"

	// TwitterAPIからアクセストークンの取得
	twitterAPIResponse, err := oauthClient.Get(nil, tokenCard, "https://api.twitter.com/1.1/account/verify_credentials.json", nil)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(twitterAPIResponse.Body)

	// 取得したユーザー情報をJSONから変換する
	var user Account
	err = json.Unmarshal(body, &user)

	// Sessionテーブル
	sessionTable := db.Table("session")
	// Userテーブル
	userTable := db.Table("user")

	// session_idの作成
	id := createSessionID(user.ScreenName)

	// Sessionテーブルに格納するレコードの作成
	s := Session{
		ID:           id,
		AccessToken:  tokenCard.Token,
		SecretToken:  tokenCard.Secret,
		RegisterDate: time.Now().Format(format),
		ScreenName:   user.ScreenName,
		UserID:       user.ID,
	}

	// SessionへのINSERTの実行
	if err = sessionTable.Put(s).Run(); err != nil {
		fmt.Println("err")
		panic(err.Error())
	}

	// UserへのINSERTの実行
	if err = userTable.Put(user).Run(); err != nil {
		fmt.Println("err")
		panic(err.Error())
	}

	jsonBytes, _ := json.Marshal(s)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "origin,Accept,Authorization,Content-Type",
			"Content-Type":                 "application/json",
		},
		Body: string(jsonBytes),
	}, nil

}

func main() {
	lambda.Start(handler)
}
