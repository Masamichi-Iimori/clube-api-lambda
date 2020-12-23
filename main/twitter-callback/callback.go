package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
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
	Cookie        string `json:"Cookie"`
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
	ID           string `dynamo:"id"`
	AccessToken  string `dynamo:"access_token"`
	SecretToken  string `dynamo:"secret_token"`
	RegisterDate string `dynamo:"register_date"`
}

// Account TwitterAPIから取得したユーザー情報から、screen_nameを取り出すための構造体
type Account struct {
	ScreenName string `json:"screen_name"`
}

// Response APIGatewayへのレスポンスを定義するための構造体
type Response struct {
	Location string `json:"location"`
	Cookie   string `json:"cookie"`
}

// Cookie クッキー
type Cookie struct {
	ID string `json:id`
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

	// session_idの取り出し
	assigned := regexp.MustCompile("id=(.*)")
	group := assigned.FindSubmatch([]byte(request.Headers["Cookie"]))
	sessionID := string(group[1])

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
	response, err := oauthClient.Get(nil, tokenCard, "https://api.twitter.com/1.1/account/verify_credentials.json", nil)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(response.Body)

	// 取得したユーザー情報をJSONから変換する
	var user Account
	err = json.Unmarshal(body, &user)

	// Sessionテーブル
	sessionTable := db.Table("session")

	// session_idの作成
	id := createSessionID(user.ScreenName)

	// Sessionテーブルに格納するレコードの作成
	s := Session{
		ID:           id,
		AccessToken:  tokenCard.Token,
		SecretToken:  tokenCard.Secret,
		RegisterDate: time.Now().Format(format),
	}
	// INSERTの実行
	if err = sessionTable.Put(s).Run(); err != nil {
		fmt.Println("err")
		panic(err.Error())
	}

	// リダイレクトさせてCookieをつけたい
	redirectURL := "https://clubhub.ga"

	// returnResponse := Response{
	// 	Location: redirectURL,
	// 	Cookie:   fmt.Sprintf("id=%s;", id),
	// }

	// jsonBytes, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: 301,
		Headers: map[string]string{
			"Location":   redirectURL,
			"Set-Cookie": fmt.Sprintf("id=%s;", id),
		},
	}, nil

}

func main() {
	lambda.Start(handler)
}
