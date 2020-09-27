.PHONY: build

build:
	sam build

# 追加
packege:
	sam package --template-file template.yaml --output-template-file output-template.yaml --s3-bucket clubes-api-go --profile masamichi

# 追加
deploy:
	#sam deploy --template-file output-template.yaml --stack-name clubes-api-go --capabilities CAPABILITY_IAM --profile masamichi
	sam deploy

delete:
	aws cloudformation delete-stack --stack-name clubes-api-go
