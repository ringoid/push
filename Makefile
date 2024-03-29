stage-all: clean stage-deploy
test-all: clean test-deploy
prod-all: clean prod-deploy

build:
	@echo '--- Building update-token function ---'
	GOOS=linux go build update-token/update.go
	@echo '--- Building update-fcm-token function ---'
	GOOS=linux go build update-fcm-token/update_fcm_token.go
	@echo '--- Building test-publish function ---'
	GOOS=linux go build test-publish/publish.go
	@echo '--- Building scheduler-publish function ---'
	GOOS=linux go build push-scheduler/scheduler.go
	@echo '--- Building internal-handle-task function ---'
	GOOS=linux go build lambda-handle-task/internal_handle_task.go lambda-handle-task/special_push.go
	@echo '--- Building internal-handle-stream function ---'
	GOOS=linux go build lambda-handle-stream/handle_stream.go lambda-handle-stream/delete_user.go

zip_lambda: build
	@echo '--- Zip update-token function ---'
	zip update.zip ./update
	@echo '--- Zip update-fcm-token function ---'
	zip update_fcm_token.zip ./update_fcm_token
	@echo '--- Zip test-publish function ---'
	zip publish.zip ./publish
	@echo '--- Zip scheduler-publish function ---'
	zip scheduler.zip ./scheduler
	@echo '--- Zip internal-handle-task function ---'
	zip internal_handle_task.zip ./internal_handle_task
	@echo '--- Zip internal-handle-stream function ---'
	zip handle_stream.zip ./handle_stream

test-deploy: zip_lambda
	@echo '--- Build lambda test ---'
	@echo 'Package template'
	sam package --template-file push-template.yaml --s3-bucket ringoid-cloudformation-template --output-template-file push-template-packaged.yaml
	@echo 'Deploy test-push-stack'
	sam deploy --template-file push-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name test-push-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=test --no-fail-on-empty-changeset

stage-deploy: zip_lambda
	@echo '--- Build lambda stage ---'
	@echo 'Package template'
	sam package --template-file push-template.yaml --s3-bucket ringoid-cloudformation-template --output-template-file push-template-packaged.yaml
	@echo 'Deploy stage-push-stack'
	sam deploy --template-file push-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name stage-push-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=stage --no-fail-on-empty-changeset

prod-deploy: zip_lambda
	@echo '--- Build lambda prod ---'
	@echo 'Package template'
	sam package --template-file push-template.yaml --s3-bucket ringoid-cloudformation-template --output-template-file push-template-packaged.yaml
	@echo 'Deploy prod-push-stack'
	sam deploy --template-file push-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name prod-push-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=prod --no-fail-on-empty-changeset

clean:
	@echo '--- Delete old artifacts ---'
	rm -rf update
	rm -rf update.zip
	rm -rf push-template-packaged.yaml
	rm -rf publish
	rm -rf publish.zip
	rm -rf scheduler
	rm -rf scheduler.zip
	rm -rf internal_handle_task
	rm -rf internal_handle_task.zip
	rm -rf handle_stream
	rm -rf handle_stream.zip
	rm -rf update_fcm_token
	rm -rf update_fcm_token.zip

