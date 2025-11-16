import json
from enum import StrEnum
from typing import Optional

from pydantic import BaseModel


class SubscriptionType(StrEnum):
    RENEWAL = "subscription.renewed"
    CREATED = "subscription.created"
    CANCELLED = "subscription.cancelled"


class SupportedMethods(StrEnum):
    GET = "GET"
    POST = "POST"


class UserParamsSchema(BaseModel):
    userId: str


class EventSchema(BaseModel):
    httpMethod: str
    path: str
    body: Optional[str] = None
    pathParameters: Optional[UserParamsSchema] = None

    @property
    def is_get(self) -> bool:
        return self.httpMethod == SupportedMethods.GET

    @property
    def is_post(self) -> bool:
        return self.httpMethod == SupportedMethods.POST

    def parse_body(self) -> dict:
        if self.body:
            return json.loads(self.body)
        return {}


class MetadataSchema(BaseModel):
    planSku: str
    autoRenew: bool
    paymentMethod: str
    cancelReason: Optional[str] = None


class SubscriptionEventPayload(BaseModel):
    eventId: str
    eventType: str
    timestamp: str
    provider: str
    subscriptionId: str
    paymentId: Optional[str] = None
    userId: str
    customerId: str
    expiresAt: str
    cancelledAt: Optional[str] = None
    metadata: MetadataSchema

    @property
    def sub_pk(self) -> str:
        return f"user:{self.userId}"

    @property
    def sub_sk(self) -> str:
        return f"sub:{self.subscriptionId}"

    @property
    def plan_pk(self) -> str:
        return f"{self.metadata.planSku}"

    @property
    def plan_name(self) -> str:
        return self.metadata.planSku.replace("_", " ").title()

    @property
    def is_renewal(self) -> bool:
        return self.eventType == SubscriptionType.RENEWAL

    @property
    def is_created(self) -> bool:
        return self.eventType == SubscriptionType.CREATED

    @property
    def is_cancelled(self) -> bool:
        return self.eventType == SubscriptionType.CANCELLED
