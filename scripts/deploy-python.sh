#!/bin/bash

mkdir .temp
mkdir .temp/package

pip install -r app/python/requirements.txt -t .temp/package
cp -r app/python/src/. .temp/package

cd .temp/package
zip -r ../package.zip .
cd ../..

aws lambda update-function-configuration \
    --profile "fender" \
    --function-name fender_digital_code_exercise \
    --runtime python3.13 \
    --handler main.handler

aws lambda update-function-code \
    --profile "fender" \
    --function-name fender_digital_code_exercise \
    --zip-file fileb://.temp/package.zip 

rm -rf .temp