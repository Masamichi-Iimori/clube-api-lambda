package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strconv"
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

// Account アカウント
type Account struct {
	ID              string `json:"id_str"`
	ScreenName      string `json:"screen_name"`
	ProfileImageURL string `json:"profile_image_url"`
	Email           string `json:"email"`
}

type TwitterRequest struct {
	RequestToken       string `json:"request_token"`
	RequestTokenSecret string `json:"request_token_secret"`
	URL                string `json:"url"`
}

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

type Token struct {
	Id           string `dynamo:"id"`
	OauthToken   string `dynamo:"oauth_token"`
	SecretToken  string `dynamo:"secret_token"`
	RegisterDate string `dynamo:"register_date"`
}

// APIGatewayへのレスポンスを定義するための構造体
type Response struct {
	Location string `json:"location"`
	Cookie   string `json:"cookie"`
}

// GetConnect 接続を取得する
func GetConnect() *oauth.Client {
	return &oauth.Client{
		TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authorize",
		TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
		Credentials: oauth.Credentials{
			Token:  os.Getenv("CONSUMER_KEY"),
			Secret: os.Getenv("CONSUMER_SECRET"),
		},
	}
}

// GetAccessToken アクセストークンを取得する
func GetAccessToken(rt *oauth.Credentials, oauthVerifier string) (*oauth.Credentials, error) {
	oc := GetConnect()
	at, _, err := oc.RequestToken(nil, rt, oauthVerifier)

	return at, err
}

// GetMe 自身を取得する
func GetMe(at *oauth.Credentials, user *Account) error {
	oc := GetConnect()

	v := url.Values{}
	v.Set("include_email", "true")

	resp, err := oc.Get(nil, at, "https://api.twitter.com/1.1/account/verify_credentials.json", v)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return errors.New("Twitter is unavailable")
	}

	if resp.StatusCode >= 400 {
		return errors.New("Twitter request is invalid")
	}

	err = json.NewDecoder(resp.Body).Decode(user)
	if err != nil {
		return err
	}

	return nil

}

func createSessionID() string {
	str := strconv.Itoa(rand.Intn(1000)) + time.Now().Format("2006-01-02 15:04:05")
	return base64.URLEncoding.EncodeToString([]byte(str))
}

func handler() (events.APIGatewayProxyResponse, error) {

	config := GetConnect()

	callbackURL := "http://127.0.0.1:3000/twitter/callback"

	tempCredentials, err := config.RequestTemporaryCredentials(nil, callbackURL, nil)
	if err != nil {
		panic(err)
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

	// 時間取得時のフォーマット指定
	format := "2006-01-02 15:04:05"

	// session_idの作成
	id := createSessionID()

	t := Token{
		Id:           id,
		OauthToken:   tempCredentials.Token,
		SecretToken:  tempCredentials.Secret,
		RegisterDate: time.Now().Format(format),
	}

	if err := tokenTable.Put(t).Run(); err != nil {
		fmt.Println("err")
		panic(err.Error())
	}

	authorizeURL := config.AuthorizationURL(tempCredentials, nil)

	response := Response{
		Location: authorizeURL,
		Cookie:   fmt.Sprintf("id=%s;", id),
	}

	jsonBytes, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: 301,
		Headers: map[string]string{
			"Location":   authorizeURL,
			"Set-Cookie": fmt.Sprintf("id=%s;", id),
		},
		Body: string(jsonBytes),
	}, nil

}

func main() {
	lambda.Start(handler)
}
