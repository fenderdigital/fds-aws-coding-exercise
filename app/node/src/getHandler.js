
import { QueryCommand } from '@aws-sdk/lib-dynamodb'
import { ddb, TABLE_NAME, createResponse } from '.'

export const getHandler = async()=>{
    try{
        const subscriptionQueryParams = {
            TableName: TABLE_NAME,
            KeyConditionExpression: "pk = :pk AND begins_with(sk, :skPrefix)",
            ExpressionAttributeValues:{
                ":pk": `user:${userId}`,
                ":skPrefix": "sub:"
            }
        }
        const subscriptionResponse = await ddb.send(new QueryCommand(subscriptionQueryParams))
        const subscription = subscriptionResponse.Items[0]
        const planQueryParams = {
            TableName: TABLE_NAME,
            KeyConditionExpression: "pk = :pk",
            ExpressionAttributeValues:{
                ":pk": `${subscription.planSku}`
            }
        }
        const planResponse = await ddb.send(new QueryCommand(planQueryParams))
        const plan = planResponse.Items[0]
        const response = {
            userId,
            subscriptionId: subscription.pk,
            plan:{
                sku: planSku,
                name: plan.name,
                price: plan.price,
                currency: plan.currency,
                billingCycle: plan.billingCycle,
                features: plan.features
            },
            startDate: subscription.startDate,
            expiresAt: subscription.expiresAt,
            cancelledAt: subscription.cancelledAt,
            status: subscription.status,
            attributes: subscription.attributes
        }
        return createResponse(200, response)
    }catch(error){
        return createResponse(500, {message: 'Internal server error'})
    }
}