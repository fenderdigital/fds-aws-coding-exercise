# 🎸 Fender Digital - 🖥️ Interview Coding Exercise - ☁️ AWS

## 🌐 Overview
You are developing the backend system for a music streaming platform. This streaming platform will be based on subscriptions and plans where customers can select which plan they want to subscribe to and receive the benefits from that plan.

The backend system you are going to develop is designed around a serverless architecture using the Amazon Web Services platform. It is composed of an API Gateway, connected to a Lambda function which uses DynamoDB as the database.

The system should be able to support two use cases.
- Getting the subscription data for a user
- Handle incoming subscription webhook events for creation, renewal and cancellation

## 📝 Task
- Configure an API Gateway REST API to expose the following endpoints, wired to a single Lambda function
    - `GET /api/v1/subscriptions/{userId}`
    - `POST /api/v1/webhooks/subscriptions`

- Write the code for the Lambda function to handle both operations

## 🎯 Technical requirements
- Each user can only have one active subscription at a time
- The subscription `status` field must be derived from the data using the following rules:
    - The status is ACTIVE if the `canceledAt` field is not set
    - The status is PENDING if the `canceledAt` field is set, but current date is before the `expiresAt` date
    - The status is CANCELED if the `canceledAt` field is set, and current date is after the `expiresAt` date
- Subscriptions cannot be created for plans with `INACTIVE` status

## 📌 Sample objects
`GET /api/v1/subscriptions/{userId}` response
```
{
  "userId": "123",
  "subscriptionId": "sub_456789",
  "sku": "PREMIUM_MONTHLY",
  "name": "Premium Monthly",
  "price": 9.99,
  "currency": "USD",
  "billingCycle": "MONTHLY",
  "features": [
    "HD Streaming",
    "Offline Downloads",
    "Ad Free"
  ],
  "startDate": "2024-03-20T10:00:00Z",
  "expiryDate": "2024-04-20T10:00:00Z",
  "cancelledAt": null,
  "status": "ACTIVE",
  "attributes": {
    "autoRenew": true,
    "paymentMethod": "CREDIT_CARD"
  }
}
```

`POST /api/v1/webhooks/subscriptions` request body
- Subscription creation event (`subscription.created`)
```json
{
  "eventId": "evt_123456789",
  "eventType": "subscription.created",
  "timestamp": "2024-03-20T10:00:00Z",
  "provider": "STRIPE",
  "subscriptionId": "sub_456789",
  "paymentId": "pm_123456",
  "userId": "123",
  "customerId": "cus_789012",
  "expiryDate": "2024-04-20T10:00:00Z",
  "metadata": {
    "planSku": "PREMIUM_MONTHLY",
    "autoRenew": true,
    "paymentMethod": "CREDIT_CARD"
  }
}
```

- Subscription renewal event (`subscription.renewed`)
```json
{
  "eventId": "evt_987654321",
  "eventType": "subscription.renewed",
  "timestamp": "2024-04-20T10:00:00Z",
  "provider": "STRIPE",
  "subscriptionId": "sub_456789",
  "paymentId": "pm_654321",
  "userId": "123",
  "customerId": "cus_789012",
  "expiryDate": "2024-05-20T10:00:00Z",
  "metadata": {
    "planSku": "PREMIUM_MONTHLY",
    "autoRenew": true,
    "paymentMethod": "CREDIT_CARD"
  }
}
```

- Subscription cancelation event (`subscription.canceled`)
```json
{
  "eventId": "evt_456789123",
  "eventType": "subscription.canceled",
  "timestamp": "2024-05-20T10:00:00Z",
  "provider": "STRIPE",
  "subscriptionId": "sub_456789",
  "paymentId": null,
  "userId": "123",
  "customerId": "cus_789012",
  "expiryDate": "2024-05-20T10:00:00Z",
  "cancelledAt": "2024-05-20T10:00:00Z",
  "metadata": {
    "planSku": "PREMIUM_MONTHLY",
    "autoRenew": false,
    "paymentMethod": "CREDIT_CARD",
    "cancelReason": "USER_REQUESTED"
  }
}
```

## ⚙️ Setting up
### Configuring the AWS CLI

Some steps in the coding exercise process require interaction with AWS through the AWS CLI. 

If you do not have the AWS CLI installed, follow the [tutorial](https://docs.aws.amazon.com/cli/v1/userguide/cli-chap-install.html)

You will need to create a new profile called `fender`. You can do so by running the following command and entering the variables.

```bash
aws configure --profile fender
```

```
AWS Access Key ID: <your access key id>
AWS Secret Access Key: <your secret access key>
Default region name: us-east-1
Default output format: json
```

To ensure correct configuration, run the following command

```bash
aws lambda list-functions --profile fender
```

You should see a function called `fender_digital_code_exercise`

### Environment variables

To manage environment variables, create a `.env` file in the root directory of the repository. This file will be used to sync the Lambda runtime environment variables when deployed.

The `.env` file MUST follow the traditional convention of KEY=VALUE in order for the deployment to work. 

Here's an example:

```sh
VARIABLE_ONE=Hello
VARIABLE_TWO=World!

# Comments and spaces are allowed!
ANOTHER_ONE=foo

LAST_ONE=bar
```

## 🚀 Development and deployment
### Developing your solution
We provide detailed instructions on development for the following languages.

- [Python](/app/python/readme.md)
- [Node.js](/app/node/readme.md)
- [Go](/app/go/readme.md)

If you want to use a different language, create a new folder in the `app` directory and manage it however you want.

### Deploying to AWS
The first step is to deploy any environment variables to the Lambda runtime.
Make sure you have a `.env` file in the repository root and run the following command.

```sh
make deploy-env
```

This will take all of the variables in the `.env` file you created and add them to the Lambda runtime.

If you used one of the supported languages, run one of the following commands to deploy your code. 
These will only work if the development instructions for each language were followed and the AWS CLI was correctly set up.

```sh
make deploy-node    # For Node.js runtime
make deploy-python  # For Python runtime
make deploy-go      # For Go runtime
```

If you decided to use a different language, you will have to manually configure and deploy the Lambda function.

## 🧪 Testing
To test your integration, create and configure an API Gateway stage, and call the Invoke URL with the desired path.

An E2E test should be created and should do the following operations.

- Create a new test plan (done manually)
- Call the `POST /api/v1/webhooks/subscriptions` endpoint to create a new subscription
- Call the `GET /api/v1/subscriptions/{userId}` endpoint to retrieve data for the newly created subscription
- Call the `POST /api/v1/webhooks/subscriptions` endpoint to renew the subscription
- Call the `GET /api/v1/subscriptions/{userId}` endpoint to verify the renewal
- Call the `POST /api/v1/webhooks/subscriptions` endpoint to cancel the subscription
- Call the `GET /api/v1/subscriptions/{userId}` endpoint to verify the cancelation
- Clean up all of the test data from the table (done manually)

You can use cURL commands or a tool like Postman to write your tests.

## ✅ Submitting
When the exercise is complete, send a pull request from your fork to the parent repository. 
The forked repository should contain your Lambda source code somewhere in the `app` folder.
