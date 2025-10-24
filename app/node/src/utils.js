
import { DynamoDBClient } from '@aws-sdk/client-dynamodb'
import { DynamoDBDocumentClient } from '@aws-sdk/lib-dynamodb'

const config = {
    region: process.env.REGION,
}
const client = new DynamoDBClient(config)
export const createResponse = (statusCode, body) => {
    return {
        statusCode, body: JSON.stringify(body), headers: { "Content-Type": "application/json" },
    }
}

export const ddb = DynamoDBDocumentClient.from(client, {marshallOptions: { convertClassInstanceToMap: true, removeUndefinedValues: true}})

export const TABLE_NAME = process.env.TABLE_NAME

export const billingCycles = Object.freeze({
    MONTHLY: 'MONTHLY',
    YEARLY: 'YEARLY'
})