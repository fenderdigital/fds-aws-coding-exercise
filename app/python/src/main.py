import json
import os
from datetime import datetime, timezone

import boto3
from boto3.dynamodb.conditions import Key

ddb = boto3.resource('dynamodb')
table = ddb.Table(os.environ['AWS_DDB_TABLE_NAME'])

def handler(event, context):
    match event['resource']:
        case '/api/v1/subscriptions/{userId}':
            match event['httpMethod']:
                case 'GET':
                    return handle_get_subscription(event)
                
        case '/api/v1/webhooks/subscriptions':
            match event['httpMethod']:
                case 'POST':
                    return handle_subscription_webhooks(event)
                
def build_response(status_code, body = None):
    response =  {
        'statusCode': status_code,
        'headers': {'Content-Type': 'application/json'},
    }
    
    if body:
        response['body'] = json.dumps(body)

    return response

def handle_get_subscription(event):
    userId = event['pathParameters']['userId']

    response = table.query(
        KeyConditionExpression=(
            Key('pk').eq(f'user:{userId}')
            & Key('sk').begins_with('sub:')
        )
    )

    if not (sub := next(iter(response['Items']), None)):
        return build_response(404, {'error': f'Subscription not found for user {userId}'})

    plan = table.get_item(Key={'pk': f'plan:{sub["planSku"]}', 'sk': 'meta'}).get('Item')

    status = 'active'
    if sub['cancelledAt']:
        if datetime.fromisoformat(sub['expiresAt']) > datetime.now(timezone.utc):
            status = 'pending'
        else:
            status = 'cancelled'

    output = {
        'userId': userId,
        'subscriptionId': sub['sk'].split(':')[1],
        'sku': sub['planSku'],
        'name': plan['name'],
        'price': float(plan['price']),
        'currency': plan['currency'],
        'billingCycle': plan['billingCycle'],
        'features': plan['features'],
        'startDate': sub['startDate'],
        'expiresAt': sub['expiresAt'],
        'cancelledAt': sub.get('cancelledAt'),
        'status': status,
        'attributes': sub.get('attributes', {})
    }

    return build_response(200, output)

def handle_subscription_webhooks(event):
    payload = json.loads(event['body'])
    try:
        match payload['eventType']:
            case 'subscription.created':
                return handle_subscription_creation(payload)
            case 'subscription.renewed':
                return handle_subscription_renewal(payload)
            case 'subscription.cancelled':
                return handle_subscription_cancellation(payload)
            case _:
                return build_response(400, {'error': 'Unsupported event type'})
    except KeyError as e:
        return build_response(400, {'error': f'Missing required field: {e}'})

def handle_subscription_creation(payload):
    result = table.query(
        KeyConditionExpression=(
            Key('pk').eq(f'user:{payload["userId"]}')
            & Key('sk').begins_with('sub:')
        )
    )

    if result['Count'] > 0:
        return build_response(400, {'error': 'User already has an active subscription'})

    plan = table.get_item(Key={'pk': f'plan:{payload["metadata"]["planSku"]}', 'sk': 'meta'}).get('Item')
    if not plan or plan['status'] != 'active':
        return build_response(400, {'error': 'Plan not found or inactive'})
    
    item = {
        'pk': f'user:{payload["userId"]}',
        'sk': f'sub:{payload["subscriptionId"]}',
        'type': 'sub',
        'planSku': payload['metadata']['planSku'],
        'startDate': payload['timestamp'],
        'expiresAt': payload['expiresAt'],
        'lastModifiedAt': payload['timestamp'],
        'attributes': payload.get('metadata', None)
    }

    table.put_item(Item=item)

    return build_response(204)

def handle_subscription_renewal(payload):
    table.update_item(
        Key={
            'pk': f'user:{payload["userId"]}',
            'sk': f'sub:{payload["subscriptionId"]}'
        },
        UpdateExpression="SET #expiresAt = :expiresAt, #attributes = :attributes, #lastModifiedAt = :lastModifiedAt",
        ExpressionAttributeNames={
            '#expiresAt': 'expiresAt',
            '#attributes': 'attributes',
            '#lastModifiedAt': 'lastModifiedAt'
        },
        ExpressionAttributeValues={
            ':expiresAt': payload['expiresAt'],
            ':attributes': payload.get('metadata', None),
            ':lastModifiedAt': payload['timestamp']
        }
    )

    return build_response(204)

def handle_subscription_cancellation(payload):
    table.update_item(
        Key={
            'pk': f'user:{payload["userId"]}',
            'sk': f'sub:{payload["subscriptionId"]}'
        },
        UpdateExpression="SET #cancelledAt = :cancelledAt, #attributes = :attributes, #lastModifiedAt = :lastModifiedAt",
        ExpressionAttributeNames={
            '#cancelledAt': 'cancelledAt', 
            '#attributes': 'attributes',
            '#lastModifiedAt': 'lastModifiedAt'
        },
        ExpressionAttributeValues={
            ':cancelledAt': payload['cancelledAt'],
            ':attributes': payload.get('metadata', None),
            ':lastModifiedAt': payload['timestamp']
        }
    )

    return build_response(204)