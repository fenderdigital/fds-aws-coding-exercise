#!/bin/bash

mkdir .temp
mkdir .temp/package

cd app/node
npm install
cd ../..

cp -r app/node/node_modules .temp/package/
cp -r app/node/src/. .temp/package/

cd .temp/package
zip -r ../package.zip .
cd ../..

aws lambda update-function-configuration \
    --profile "fender" \
    --function-name fender_digital_code_exercise \
    --runtime nodejs22.x \
    --handler index.handler

aws lambda update-function-code \
    --profile "fender" \
    --function-name fender_digital_code_exercise \
    --zip-file fileb://.temp/package.zip

rm -rf .temp