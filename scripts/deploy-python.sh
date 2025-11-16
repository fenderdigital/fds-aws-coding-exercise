#!/bin/bash

mkdir .temp
mkdir .temp/package

# User `uv` for faster installs
uv pip install -r app/python/requirements.txt --target .temp/package
#! NOTE: Since we're using pydantic v2, please watch step in how to deploy
#! pydantic with AWS Lambda: 
#! https://docs.pydantic.dev/latest/integrations/aws_lambda/#installing-pydantic-for-aws-lambda-functions
#! Commente out if not using pydantic v2
pip install \
  --platform manylinux2014_x86_64 \
  --implementation cp \
  --python-version 3.13 \
  --only-binary=:all: \
  --target .temp/package \
  --upgrade pydantic

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