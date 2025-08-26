# 🎸 Fender Digital • AWS Coding Exercise ☁️

## 🌐 Overview
You are developing the backend system for a music streaming platform. 
This streaming platform will be based on subscriptions and plans where customers can select which plan they want to subscribe to.

The backend system you are going to develop is designed around a serverless architecture using the Amazon Web Services (AWS) platform. 
It is composed of an API Gateway with multiple endpoints, connected to a Lambda function which uses DynamoDB as the database.

| ![arch.svg](img/arch.svg)    | 
| :--:                         | 
| *Cloud architecture diagram* |

The system should be able to support two use cases.
- Getting the subscription data for a user
- Handle incoming subscription webhook events for creation, renewal and cancellation

## 📝 Task
- Add the following endpoints to the `fender_digital_code_exercise` REST API using the API Gateway service in the AWS Console
    - `GET /api/v1/subscriptions/{userId}`
    - `POST /api/v1/webhooks/subscriptions`

- Create a Lambda proxy integration for the endpoints which calls the `fender_digital_code_exercise` Lambda function

- Create a deployment and stage for the `fender_digital_code_exercise` REST API

- Create an API key and usage plan for the `fender_digital_code_exercise` REST API and connect it to the created stage

- Write the code for the Lambda function to handle both operations

- Deploy the code from your local environment to AWS using the provided deployment tools

- Write end-to-end (E2E) tests for the subscription flow

## 🎯 Technical requirements
- Each user can only have one active subscription at a time

- The subscription `status` field must be derived from the data using the following rules:

    - The status is `ACTIVE` if the `canceledAt` field is not set
    - The status is `PENDING` if the `canceledAt` field is set, but current date is before the `expiresAt` date
    - The status is `CANCELED` if the `canceledAt` field is set, and current date is after the `expiresAt` date

- Subscriptions cannot be created for plans with `INACTIVE` status

- `PLAN` items must be created manually in the AWS Console DynamoDB service

- The response of the `GET /api/v1/subscriptions/{userId}` should also contain the associated plan data

## 📌 Sample objects

### `sub` DynamoDB item

| Field name     | Description                                          | DynamoDB type |
| :--            | :--                                                  | :--           |
| `pk`           | Partition key of the item (e.g. `user:<userId>`)     | `String`      |
| `sk`           | Sort key of the item (e.g. `sub:<subId>}`)           | `String`      |
| `type`         | Item type (always `sub`)                             | `String`      |
| `planSku`      | SKU of the subscription plan                         | `String`      |
| `startDate`    | ISO-8601 string of subscription start datetime       | `Number`      |
| `expiresAt`    | ISO-8601 string of subscription expiration datetime  | `String`      |
| `canceledAt`   | ISO-8601 string of subscription cancelation datetime | `String`      |
| `lastModified` | ISO-8601 string of last modified datetime            | `String`      |
| `attributes `  | Extra attributes for the subscription (metadata)     | `Map`         |

### `plan` DynamoDB item

| Field name     | Description                                      | DynamoDB type |
| :--            | :--                                              | :--           |
| `pk`           | Partition key of the item                        | `String`      |
| `sk`           | Sort key of the item                             | `String`      |
| `type`         | Item type (always `plan`)                        | `String`      |
| `name`         | Name of the plan                                 | `String`      |
| `price`        | Price of the plan                                | `Number`      |
| `currency`     | Currency of the plan price                       | `String`      |
| `billingCycle` | Billing cycle of the plan (`monthy` or `yearly`) | `String`      |
| `features`     | List of features (as strings)                    | `List`        |
| `status`       | Status of the plan (`active` or `inactive`)      | `String`      |
| `lastModified` | ISO-8601 string of last modified datetime        | `String`      |

### `GET /api/v1/subscriptions/{userId}` response

```json
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
  "expiresAt": "2024-04-20T10:00:00Z",
  "cancelledAt": null,
  "status": "ACTIVE",
  "attributes": {
    "autoRenew": true,
    "paymentMethod": "CREDIT_CARD"
  }
}
```

### `POST /api/v1/webhooks/subscriptions` request body

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
  "expiresAt": "2024-04-20T10:00:00Z",
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
  "expiresAt": "2024-05-20T10:00:00Z",
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
  "expiresAt": "2024-05-20T10:00:00Z",
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
### ⏮️ Prerequisites

- A Unix-based OS (Linux distro, MacOS or WSL2)
- AWS CLI v2 ([installation guide](https://docs.aws.amazon.com/cli/v1/userguide/cli-chap-install.html))
### ☁️ Configuring the AWS CLI

Some steps in the coding exercise process require interaction with AWS through the AWS CLI. You will need to create a new profile called `fender`. You can do so by running the following command and entering the variables.

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

### 🌳 Environment variables

To manage environment variables, create a `.env` file in the root directory of the repository. This file will be used to sync the Lambda runtime environment variables when deployed.

The `.env` file MUST follow the traditional convention of `KEY=VALUE` in order for the deployment to work. Here's an example:

```sh
VARIABLE_ONE=Hello
VARIABLE_TWO=World!

# Comments and newlines are allowed
ANOTHER_ONE=foo

LAST_ONE=bar
```

## 🖥️ Development and deployment
### 🧠 Developing your solution
We provide detailed instructions on development for the following languages.

- [Python](/app/python/readme.md)
- [Node.js](/app/node/readme.md)
- [Go](/app/go/readme.md)

If you want to use a different language, create a new folder in the `app` directory and manage it however you want.

### 🚀 Deploying to AWS
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
