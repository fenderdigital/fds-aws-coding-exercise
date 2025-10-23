import { DynamoDBClient } from '@aws-sdk/client-dynamodb'
import { DynamoDBDocumentClient, QueryCommand } from '@aws-sdk/lib-dynamodb'

const client = new DynamoDBClient({
    region: process.env.REGION
})

const ddb = DynamoDBDocumentClient.from(client, {marshallOptions: { convertClassInstanceToMap: true, removeUndefinedValues: true}})

const TABLE_NAME = process.env.TABLE_NAME
const createResponse = (statusCode, body) => ({
    statusCode, body, headers: { "Content-Type": "application/json" },
})

exports.handler = async (event) => {
    const userId = event?.pathParameters?.userId
    if(!userId){
        return createResponse(
            400, {message: 'Missing UserId!'}
        )
    }
    switch(event.httpMethod){
        case 'GET':
            try{
                const params = {
                    TableName: TABLE_NAME,
                    KeyConditionExpression: "pk = :pk AND begins_with(sk, :skPrefix)",
                    ExpressionAttributeValues:{
                        ":pk": `user:${userId}`,
                        ":skPrefix": "sub:"
                    }
                }
                const result = await ddb.send(new QueryCommand(params))
                return createResponse(200, result)
            }catch(error){
                return createResponse(500, {message: 'Internal server error'})
            }
        case 'POST':
            
            return
        default:
            return createResponse(405, {message: 'Method not allowed!'})
    }
};

