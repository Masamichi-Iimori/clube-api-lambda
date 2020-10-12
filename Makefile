include .env


$(eval export $(shell sed -ne 's/ *#.*$//; /./ s/=.*$$// p' .env))

.PHONY: build

build:
	sam build

# 追加
packege:
	sam package --template-file template.yaml --output-template-file output-template.yaml --s3-bucket clubes-api-go --profile masamichi

# 追加
deploy:
	#sam deploy --template-file output-template.yaml --stack-name clubes-api-go --capabilities CAPABILITY_IAM --profile masamichi
	sam deploy --parameter-overrides ConsumerKey=${CONSUMER_KEY} ConsumerSecret=${CONSUMER_SECRET} AccessToken=${ACCESS_TOKEN} AccessTokenSecret=${ACCESS_TOKEN_SECRET}

delete:
	aws cloudformation delete-stack --stack-name clubes-api-go

local:
	sam local start-api --parameter-overrides ConsumerKey=${CONSUMER_KEY} ConsumerSecret=${CONSUMER_SECRET} AccessToken=${ACCESS_TOKEN} AccessTokenSecret=${ACCESS_TOKEN_SECRET}
