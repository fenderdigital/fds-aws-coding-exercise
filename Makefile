
deploy-env:
	scripts/deploy-env.sh > /dev/null
	
deploy-go:
	scripts/deploy-go.sh > /dev/null

deploy-python:
	scripts/deploy-python.sh > /dev/null

deploy-node:
	scripts/deploy-node.sh > /dev/null

set_env_variables:
	@echo "Setting AWS Credentials..."
	export $(grep -v '^#' .env)

# Python only
export_serverless_requirements:
	@echo "Exporting requirements..."
	uv pip compile pyproject.toml -o app/python/requirements.txt

check_lambda_status:
	@echo "Checking Lambda function status..."
	aws lambda get-function-configuration \
		--profile "fender" \
		--function-name fender_digital_code_exercise \
		--query '{State:State, LastUpdateStatus:LastUpdateStatus, LastUpdateStatusReason:LastUpdateStatusReason}'