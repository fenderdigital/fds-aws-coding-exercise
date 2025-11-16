from http import HTTPStatus

from pydantic import BaseModel

try:
    # For local development
    from .models.models import process_subscription_and_plan, process_user_id
    from .schemas.schemas import EventSchema, SubscriptionEventPayload
    from .utils.response import error_response, validation_wrapper
except:
    # For AWS Lambda deployment
    from models.models import process_subscription_and_plan, process_user_id
    from schemas.schemas import EventSchema, SubscriptionEventPayload
    from utils.response import error_response, validation_wrapper


# /api/v1/subscriptions/{userId}
def router_get_user_subscription(user_id: str) -> dict:
    """
    Router function to handle GET /api/v1/subscriptions/{userId} requests.
    """
    return process_user_id(user_id=user_id)


# /api/v1/webhooks/subscriptions
def router_post_user_subscription(body: SubscriptionEventPayload) -> dict:
    """
    Router function to handle POST /api/v1/webhooks/subscriptions requests.
    """
    return process_subscription_and_plan(payload=body)


class Router(BaseModel):
    event: EventSchema

    @validation_wrapper
    def process_event(self) -> dict:

        if self.event.is_get:
            user_id = self.event.pathParameters.userId
            return router_get_user_subscription(user_id=user_id)

        elif self.event.is_post:
            body = SubscriptionEventPayload(**self.event.parse_body())
            return router_post_user_subscription(body=body)

        else:
            return error_response(
                "Method not allowed", status_code=HTTPStatus.METHOD_NOT_ALLOWED
            )
