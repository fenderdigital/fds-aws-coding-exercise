
import { QueryCommand } from '@aws-sdk/lib-dynamodb'
import { ddb, TABLE_NAME } from './utils.js'
import { AppError } from './error.js'

export const getUserSubscription = async(userId, subscriptionId = null, includeInactive = false)=>{
    const subscriptionQueryParams = {
        TableName: TABLE_NAME,
        KeyConditionExpression: "pk = :pk AND begins_with(sk, :skPrefix)",
        ExpressionAttributeValues:{
            ":pk": `user_${userId}`,
            ":skPrefix": `sub_`
        }
    }

    if(!includeInactive){
        subscriptionQueryParams['FilterExpression'] = "#st = :active"
        subscriptionQueryParams['ExpressionAttributeNames'] = {"#st": "status"}
        subscriptionQueryParams.ExpressionAttributeValues[":active"] =  'ACTIVE'
    }

    const subscriptionResponse = await ddb.send(new QueryCommand(subscriptionQueryParams))
    if(!subscriptionResponse.Items.length){
        return {}
    }

    const subscription = subscriptionResponse.Items[0]
    if(subscriptionId & subscriptionId !== subscription.pk){
        throw  new AppError(400, 'User subscription mismatch')
    }
    
    const plan = await getPlan(subscription.planSku)
    const response = {
        userId,
        subscriptionId: subscription.pk,
        plan:{
            sku: subscription.planSku,
            name: plan.name,
            price: plan.price,
            currency: plan.currency,
            billingCycle: plan.billingCycle,
            features: plan.features
        },
        startDate: subscription.startDate,
        expiresAt: subscription.expiresAt,
        canceledAt: subscription.canceledAt,
        status: subscription.status,
        attributes: subscription.attributes
    }
    return response
}

export const getPlan = async (planSku)=>{
    const planQueryParams = {
        TableName: TABLE_NAME,
        KeyConditionExpression: "pk = :pk",
        FilterExpression: "#st = :status",
        ExpressionAttributeNames: {"#st": "status"},
        ExpressionAttributeValues:{
            ":pk": `${planSku}`,
            ":status": 'ACTIVE'
        }
    }
    const planResponse = await ddb.send(new QueryCommand(planQueryParams))
    if(!planResponse.Items.length){
        return {}
    }
    const plan = planResponse.Items[0]
    return plan
}