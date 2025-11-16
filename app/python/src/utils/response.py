import json
from functools import wraps
from http import HTTPStatus

from pydantic import ValidationError


def success_response(body: str, data: dict = None) -> dict:

    body = {"message": body}

    if data:
        body["data"] = data

    return {
        "statusCode": HTTPStatus.OK,
        "body": json.dumps(body),
    }


def error_response(body: str, status_code: HTTPStatus = HTTPStatus.BAD_REQUEST) -> dict:
    return {
        "statusCode": status_code,
        "body": json.dumps({"error": body}),
    }


def process_pydantic_error(e):
    fields = []

    for error in e.errors():
        if error.get("loc"):
            location = error["loc"][0]
            message = error["msg"]
            error_message = f"'{location}': `{message}`"
            fields.append(error_message)
        else:
            fields.append(error["msg"])

    fields_error = f", ".join(fields)

    return fields_error


def validation_wrapper(f):
    """
    Decorator to catch validation errors
    """

    @wraps(f)
    def decorator(*args, **kwargs):
        try:
            return f(*args, **kwargs)
        except ValidationError as e:
            message = process_pydantic_error(e)
            return False, error_response(message)
        except Exception as e:
            return False, error_response(str(e))

    return decorator
