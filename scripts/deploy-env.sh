#!/bin/bash

ENV_VARS=""

while IFS='=' read -r key value || [[ -n "$key" ]]; do
    # Skip empty lines and comments
    if [[ -z "$key" || "$key" =~ ^# ]]; then
        continue
    fi

    # Build the comma-separated string
    if [[ -z "$ENV_VARS" ]]; then
        ENV_VARS="${key}=${value}"
    else
        ENV_VARS="${ENV_VARS},${key}=${value}"
    fi
done < .env

aws lambda update-function-configuration \
    --profile "fender" \
    --function-name "fender_digital_code_exercise" \
    --environment "Variables={${ENV_VARS}}"