import random
from datetime import datetime
from enum import StrEnum
from typing import Literal, Optional

from faker import Faker
from pydantic import BaseModel

try:
    # For local development
    from ..db.tables import DynamoFenderTables
    from ..schemas.schemas import SubscriptionEventPayload
    from ..utils.response import success_response, validation_wrapper
    from ..utils.utils import parse_iso8601
except:
    # For AWS Lambda deployment
    from db.tables import DynamoFenderTables
    from schemas.schemas import SubscriptionEventPayload
    from utils.response import success_response, validation_wrapper
    from utils.utils import parse_iso8601

fake = Faker("es_MX")


class SubscriptionStatus(StrEnum):
    ACTIVE = "active"
    PENDING = "pending"
    CANCELLED = "cancelled"


class BillingCycle(StrEnum):
    MONTHLY = "monthly"
    YEARLY = "yearly"


class PlanStatus(StrEnum):
    ACTIVE = "active"
    INACTIVE = "inactive"


class SubscriptionAttributes(BaseModel):
    # Adding as optional in case some attributes are missing
    provider: Optional[str] = None
    paymentId: Optional[str] = None
    customerId: Optional[str] = None
    autoRenew: Optional[bool] = None
    paymentMethod: Optional[str] = None


class SubscriptionModel(BaseModel):
    pk: str
    sk: str
    type: str = "sub"
    planSku: str
    startDate: str
    expiresAt: str
    cancelledAt: str | None = None
    lastModified: str
    attributes: SubscriptionAttributes | None = None

    @property
    def plan_pk(self) -> str:
        return f"{self.planSku}"

    def create(self) -> None:
        DynamoFenderTables.SUBSCRIPTIONS_AND_PLANS.write(
            self.model_dump(exclude_none=True)
        )

    @property
    def last_date_modified(self) -> datetime:
        return parse_iso8601(self.lastModified)

    def parse_cancelled_at(self) -> str:
        if not self.cancelledAt:
            raise ValueError("cancelledAt is None")
        return parse_iso8601(self.expiresAt)

    @property
    def is_pending(self) -> bool:
        return self.last_date_modified < self.parse_cancelled_at()

    @property
    def is_cancelled(self) -> bool:
        return self.last_date_modified >= self.parse_cancelled_at()

    def compute_status(self) -> SubscriptionStatus:
        if not self.cancelledAt:
            return SubscriptionStatus.ACTIVE
        if self.is_pending:
            return SubscriptionStatus.PENDING
        if self.is_cancelled:
            return SubscriptionStatus.CANCELLED


class PlanModel(BaseModel):
    pk: str
    sk: str
    type: str
    name: str
    price: float
    currency: str
    billingCycle: Literal[BillingCycle.MONTHLY, BillingCycle.YEARLY]
    features: list[str]
    status: Literal[PlanStatus.ACTIVE, PlanStatus.INACTIVE]
    lastModified: str | None = None

    def create(self) -> None:
        DynamoFenderTables.PLAN.write(self.model_dump())

    @property
    def is_active(self) -> bool:
        return self.status == PlanStatus.ACTIVE

    @property
    def is_inactive(self) -> bool:
        return self.status == PlanStatus.INACTIVE


class SubscriptionDetailsSchema:
    """
    Schema to transform SubscriptionEventPayload into SubscriptionModel data.
    """

    DEFAULT_TYPE = "sub"

    def __init__(self, payload: SubscriptionEventPayload) -> None:
        self.pk = payload.sub_pk
        self.sk = payload.sub_sk
        self.type = self.DEFAULT_TYPE
        self.planSku = payload.metadata.planSku
        self.startDate = payload.timestamp
        self.expiresAt = payload.expiresAt
        self.lastModified = payload.timestamp
        self.attributes = {
            "provider": payload.provider,
            "paymentId": payload.paymentId,
            "customerId": payload.customerId,
            "autoRenew": payload.metadata.autoRenew,
            "paymentMethod": payload.metadata.paymentMethod,
        }
        if payload.is_cancelled:
            self.cancelledAt = payload.cancelledAt

    def to_dict(self) -> dict:
        return self.__dict__


class SubscriptionAdapter(BaseModel):
    payload: SubscriptionEventPayload

    def get_sub_by_pk(self) -> dict:
        return DynamoFenderTables.SUBSCRIPTIONS_AND_PLANS.get_by_pk(self.payload.sub_pk)

    def get_plan_by_pk(self) -> PlanModel | None:
        if plan := DynamoFenderTables.SUBSCRIPTIONS_AND_PLANS.get_by_pk(
            self.payload.plan_pk
        ):
            return PlanModel(**plan[0])

    def create_sub(self) -> None:
        """
        Process subscription event payload and create subscription record.
        """
        data = SubscriptionDetailsSchema(self.payload).to_dict()

        subscription_model = SubscriptionModel(**data)
        subscription_model.create()

    def _update_renewal(self) -> None:
        """
        Update plan on renewal if needed.
        """
        update_dict = {
            "pk": self.payload.sub_pk,
            "sk": self.payload.sub_sk,
            "lastModified": self.payload.timestamp,
            "expiresAt": self.payload.expiresAt,
            "internalStatus": SubscriptionStatus.ACTIVE,
        }
        DynamoFenderTables.SUBSCRIPTIONS_AND_PLANS.update(update_dict)

    def _update_cancelled(self) -> None:
        """
        Update plan on renewal if needed.
        """
        update_dict = {
            "pk": self.payload.sub_pk,
            "sk": self.payload.sub_sk,
            "lastModified": self.payload.timestamp,
            "expiresAt": self.payload.expiresAt,
            "cancelledAt": self.payload.cancelledAt,
            "internalStatus": SubscriptionStatus.CANCELLED,
        }
        DynamoFenderTables.SUBSCRIPTIONS_AND_PLANS.update(update_dict)

    def process(self) -> None:
        """
        Process subscription event payload.
        """
        if not (plan := self.get_plan_by_pk()) or plan.is_inactive:
            raise ValueError("Plan is inactive or does not exist")

        if not self.get_sub_by_pk():
            return self.create_sub()

        if self.payload.is_renewal:
            # Add logic to update plan if needed on renewal
            return self._update_renewal()
        elif self.payload.is_cancelled:
            # Add logic to update plan if needed on cancellation
            return self._update_cancelled()


class PlanAdapter(BaseModel):
    payload: SubscriptionEventPayload

    def _get_by_pk(self) -> dict:
        return DynamoFenderTables.PLAN.get_by_pk(self.payload.metadata.planSku)

    def _create(self) -> None:
        """
        Process subscription event payload and create plan record.
        """
        DEFAULT_TYPE = "plan"
        DEFAULT_CURRENCY = "USD"
        # Random price and currency for demonstration purposes
        random_billing_cycle = random.choice(
            [BillingCycle.MONTHLY, BillingCycle.YEARLY]
        )
        random_features = [f"Feature {el}" for el in fake.random_sample()]
        data = {
            "pk": self.payload.metadata.planSku,
            "sk": self.payload.metadata.paymentMethod,
            "type": DEFAULT_TYPE,
            "name": self.payload.plan_name,
            "price": float(fake.numerify()),
            "currency": DEFAULT_CURRENCY,
            "billingCycle": random_billing_cycle,
            "features": random_features,
            "status": PlanStatus.ACTIVE,
            "lastModified": self.payload.timestamp,
        }
        plan_model = PlanModel(**data)
        plan_model.create()

    def process(self) -> None:
        if not self._get_by_pk():
            return self._create()


class SubscriptionAndPlanAdapter(BaseModel):
    user_id: str

    @property
    def sub_pk(self) -> str:
        return f"user:{self.user_id}"

    def _get_sub_by_pk(self) -> SubscriptionModel:
        if not (
            subscription := DynamoFenderTables.SUBSCRIPTIONS_AND_PLANS.get_by_pk(
                self.sub_pk
            )
        ):
            raise ValueError("Subscription not found")

        return SubscriptionModel(**subscription[0])

    def _get_plan_by_pk(self, plan_pk: str) -> PlanModel:
        if not (plan := DynamoFenderTables.SUBSCRIPTIONS_AND_PLANS.get_by_pk(plan_pk)):
            raise ValueError("Plan not found")

        return PlanModel(**plan[0])

    def process(self) -> dict:
        subscription = self._get_sub_by_pk()
        plan = self._get_plan_by_pk(subscription.plan_pk)

        data = {
            "userId": subscription.pk,
            "subscriptionId": subscription.sk,
            "plan": {
                "sku": plan.pk,
                "name": plan.name,
                "price": plan.price,
                "currency": plan.currency,
                "billingCycle": plan.billingCycle,
                "features": plan.features,
            },
            "startDate": subscription.startDate,
            "expiresAt": subscription.expiresAt,
            "status": subscription.compute_status(),
            "attributes": {
                "autoRenew": subscription.attributes.autoRenew,
                "paymentMethod": subscription.attributes.paymentMethod,
            },
        }

        if subscription.cancelledAt:
            data["cancelledAt"] = subscription.cancelledAt

        return data


@validation_wrapper
def process_subscription_and_plan(payload: SubscriptionEventPayload) -> None:
    subscription_adapter = SubscriptionAdapter(payload=payload)
    subscription_adapter.process()

    return success_response("Subscription and Plan processed successfully")


@validation_wrapper
def process_user_id(user_id: str) -> dict:
    """
    Fetch user subscription by user ID.
    """
    subscription_adapter = SubscriptionAndPlanAdapter(user_id=user_id)
    data = subscription_adapter.process()

    return success_response("User subscription retrieved successfully", data=data)
