
import { PutCommand, UpdateCommand } from '@aws-sdk/lib-dynamodb'
import { ddb, TABLE_NAME, createResponse, billingCycles } from './utils.js'
import { getPlan, getUserSubscription } from './getHandler.js'
import { AppError } from './error.js'

const SubEvents = Object.freeze({
    CREATED: 'subscription.created',
    RENEWED: 'subscription.renewed',
    CANCELED: 'subscription.cancelled'
})

const createSubscription = async (body)=>{
    const { userId, subscriptionId, expiresAt, metadata, timestamp } = body
    if( !userId || !subscriptionId || !expiresAt || !metadata || !timestamp ){
        throw new AppError(400, `Missing required fields`)
    }
    if( !metadata.planSku ){
        throw new AppError(400, `Missing metadata fields`)
    }

    const { planSku } = metadata

    //validate user has no subscription
    const subscription = await getUserSubscription(userId)
    
    if(Object.keys(subscription).length > 0){
        throw new AppError(400, 'User has an active subscription')
    }

    //validate requested plan is not inactive
    const plan = await getPlan(planSku)
    if(Object.keys(plan).length === 0){
        throw new AppError(400, `Requested plan don't exist or is inactive`)
    }

    //create item
    const newSubscription = {
        pk: `user_${userId}`,
        sk: `${subscriptionId}`,
        type: 'sub',
        planSku,
        startDate: timestamp,
        expiresAt,
        canceledAt: '',
        lastModifiedAt: new Date().toISOString(),
        attributes: metadata,
        status: 'ACTIVE'
    }
    await ddb.send(new PutCommand({
        TableName: TABLE_NAME,
        Item: newSubscription
    }))
    return newSubscription
}

const cancelSubscription = async(body)=>{
    const { userId, subscriptionId, expiresAt } = body

    if(!userId || !subscriptionId){
        throw new AppError(400, 'Missing required fields')
    }

    //validate subscription exists
    const subscription = await getUserSubscription(userId, subscriptionId)
    if(Object.keys(subscription).length === 0){
        throw new AppError(400, 'User has no active subscription')
    }

    //update subscription status
    const now = new Date()
    const expireDate = new Date(expiresAt)
    const updateParams = {
        TableName: TABLE_NAME,
        Key: { pk: `user_${userId}`, sk: subscriptionId },
        UpdateExpression: `
            SET canceledAt = :canceledAt,
                lastModifiedAt = :lastModifiedAt,
                #st = :status
        `,
        ExpressionAttributeNames: {"#st" : "status"},
        ExpressionAttributeValues: {
            ":canceledAt": now.toISOString(),
            ":lastModifiedAt": now.toISOString(),
            ":status": now < expireDate ? 'PENDING' : 'CANCELED'
        },
        ConditionExpression: "attribute_exists(pk) AND attribute_exists(sk)",
        ReturnValues: "ALL_NEW"
    }
    const response = await ddb.send(new UpdateCommand(updateParams))
    return response
}

const renewSubscription = async(body) =>{
    const { userId, subscriptionId, expiresAt } = body

    //validate subscription exists
    if(!userId || !subscriptionId){
        return createResponse(400, {message: 'Missing required fields'})
    }

    const subscription = await getUserSubscription(userId, subscriptionId, true)

    if(Object.keys(subscription).length === 0){
        throw new AppError(400, 'Subscription not found')
    }

    //update subscription status
    const now = new Date().toISOString()

    const updateParams = {
        TableName: TABLE_NAME,
        Key: { pk: `user_${userId}`, sk: subscriptionId },
        UpdateExpression: `
            SET lastModifiedAt = :lastModifiedAt,
                expiresAt = :expiresAt,
                canceledAt = :canceledAt,
                #st = :status
        `,
        ExpressionAttributeNames: {"#st" : "status"},
        ExpressionAttributeValues: {
            ":lastModifiedAt": now,
            ":expiresAt": expiresAt,
            ":status": 'ACTIVE',
            ":canceledAt": ''
        },
        ConditionExpression: "attribute_exists(pk) AND attribute_exists(sk)",
        ReturnValues: "ALL_NEW"
    }

    const response = await ddb.send(new UpdateCommand(updateParams))
    return response
    
}
export const postHandler = async({ eventType, parsedBody })=>{
    switch(eventType){
        case SubEvents.CREATED:
            const createdResponse = await createSubscription(parsedBody)
            return createdResponse
        case SubEvents.CANCELED:
            const canceledResponse = await cancelSubscription(parsedBody)
            return canceledResponse.Attributes
        case SubEvents.RENEWED:
            const renewedResponse = await renewSubscription(parsedBody)
            return renewedResponse.Attributes
        default:
            throw new AppError(400, 'Invalid post event')
    }
}
