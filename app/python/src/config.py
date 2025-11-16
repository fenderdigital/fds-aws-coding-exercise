import os

from dotenv import load_dotenv

load_dotenv(override=True)


class Config:
    ENVIRONMENT = os.getenv("ENVIRONMENT", "staging")
    AWS_ACCESS_KEY_ID = os.getenv("AWS_ACCESS_KEY_ID", "")
    AWS_SECRET_ACCESS_KEY = os.getenv("AWS_SECRET_ACCESS_KEY", "")
    AWS_REGION_NAME = os.getenv("AWS_REGION_NAME", "us-east-1")


IS_DEVELOPMENT = Config.ENVIRONMENT == "development"
