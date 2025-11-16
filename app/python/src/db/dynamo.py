import json
from decimal import Decimal

import boto3
from boto3.dynamodb.conditions import And, Key

try:
    # For local development
    from ..config import IS_DEVELOPMENT, Config
except:
    from config import IS_DEVELOPMENT, Config


def serialize_dynamo(dict):
    return json.loads(json.dumps(dict, default=str_dynamo_data))


def str_dynamo_data(obj):
    # Coerce every object into string
    return str(obj)


def dynamo_write_serializer(dict: dict) -> dict:
    """
    Serialize data before writing into DynamoDB
    """
    for key, value in dict.items():
        if isinstance(value, float):
            dict[key] = Decimal(str(value))

    return dict


PK_FIELD = "pk"
SK_FIELD = "sk"


class DynamoFender:
    """
    DynamoDB connection handler for Fender application.
    """

    def __init__(self, tablename) -> None:
        if IS_DEVELOPMENT:
            auth_params = {
                "aws_access_key_id": Config.AWS_ACCESS_KEY_ID,
                "aws_secret_access_key": Config.AWS_SECRET_ACCESS_KEY,
                "region_name": Config.AWS_REGION_NAME,
            }
            self.dynamodb = boto3.resource("dynamodb", **auth_params)
            self.client = boto3.client("dynamodb", **auth_params)
        else:
            self.dynamodb = boto3.resource("dynamodb")
            self.client = boto3.client("dynamodb")
        try:
            self.table = self.dynamodb.Table(tablename)
        except Exception as error:
            raise ValueError(
                f"Could't make connection to `{tablename}` table due `{error}`"
            )

    def write(self, data: list | dict) -> bool:
        """
        Writes data into table
        """

        if isinstance(data, dict):
            data = [data]

        with self.table.batch_writer() as batch:
            for values in data:
                serialized_values = dynamo_write_serializer(values)
                batch.put_item(Item=serialized_values)

        return True

    def _convert_updatable_dict(self, payload: dict) -> dict:
        """
        Convert payload to corresponding format to update
        """

        final_dict = {}

        for key, value in payload.items():
            if key == PK_FIELD or key == SK_FIELD:
                continue

            final_dict.update({key: {"Value": value}})

        return final_dict

    def update(self, data: dict) -> None:
        """
        Updates DB in dynamo
        """
        pk_value = data.get(PK_FIELD)
        sk_value = data.get(SK_FIELD)

        if PK_FIELD not in data.keys() or not pk_value:
            raise ValueError(f"`pk` is mandatory")

        if SK_FIELD not in data.keys() or not sk_value:
            raise ValueError(f"`sk` is mandatory")

        # Convert dict to value field
        attributes = self._convert_updatable_dict(data)

        return self.table.update_item(
            Key={
                PK_FIELD: pk_value,
                SK_FIELD: sk_value,
            },
            AttributeUpdates=attributes,
        )

    def get_by_pk(self, pk: str) -> dict:
        response = self.table.query(KeyConditionExpression=Key("pk").eq(pk))
        return serialize_dynamo(response["Items"])

    def get_or_create(self, pk: str, sk: str) -> dict:
        response = self.table.query(
            KeyConditionExpression=And(Key("pk").eq(pk), Key("sk").eq(sk))
        )
        items = serialize_dynamo(response["Items"])
        if items:
            return items[0]
        return {}


class SubscriptionTable(DynamoFender):
    """
    DynamoDB handler for Subscription table.
    """

    __tablename__ = "FenderSubscriptions"

    def __init__(self, tablename: str = None) -> None:
        _tablename = tablename or self.__tablename__
        super().__init__(_tablename)


class PlanTable(DynamoFender):
    """
    DynamoDB handler for Plan table.
    """

    __tablename__ = "FenderPlans"

    def __init__(self, tablename: str = None) -> None:
        _tablename = tablename or self.__tablename__
        super().__init__(_tablename)


class SubscriptionsAndPlansTable(DynamoFender):
    """
    DynamoDB handler for Plan table.
    """

    __tablename__ = "fender_digital_code_exercise"

    def __init__(self, tablename: str = None) -> None:
        _tablename = tablename or self.__tablename__
        super().__init__(_tablename)
