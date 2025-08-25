#!/bin/bash

mkdir .temp

cd app/go
go mod tidy
go mod vendor

GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o ../../.temp/bootstrap src/main.go

cd ../../.temp
zip package.zip bootstrap
cd ..

aws lambda update-function-configuration \
    --profile "fender" \
    --function-name fender_digital_code_exercise \
    --runtime provided.al2023 \
    --handler bootstrap

aws lambda update-function-code \
    --profile "fender" \
    --function-name fender_digital_code_exercise \
    --zip-file fileb://.temp/package.zip

rm -rf .temp