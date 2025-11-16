try:
    # For local development
    from .dynamo import PlanTable, SubscriptionsAndPlansTable, SubscriptionTable
except:
    # For AWS Lambda deployment
    from dynamo import PlanTable, SubscriptionsAndPlansTable, SubscriptionTable


class DynamoFenderTables:
    SUBSCRIPTION = SubscriptionTable()
    PLAN = PlanTable()
    SUBSCRIPTIONS_AND_PLANS = SubscriptionsAndPlansTable()
