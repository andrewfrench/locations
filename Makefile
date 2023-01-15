version := $(shell date -u +v%Y.%m.%d.%H%M%S)

bootstrap: ## Executes bootstrap command.
	go run cmd/bootstrap/bootstrap.go

build: build/receiver build/reducer ## Builds executables for Lambda.
	@echo "Building $(version)"

build/receiver:
	GOOS=linux go build -o bin/receiver/handler cmd/receiver/receiver.go

build/reducer:
	GOOS=linux go build -o bin/reducer/handler cmd/reducer/reducer.go

package: package/receiver package/reducer ## Builds and packages executables for Lambda.
	@echo "Packaging $(version)"

package/receiver: build/receiver
	zip -jo bin/receiver-$(version).zip  bin/receiver/handler

package/reducer: build/reducer
	zip -jo bin/reducer-$(version).zip   bin/reducer/handler

deploy/lambdas: deploy/receiver deploy/reducer ## Builds, packages, and deploys executables for Lambda.
	@echo "Deploying $(version)"

deploy/receiver: package/receiver
	aws lambda update-function-code --function-name location-receiver --zip-file fileb://bin/receiver-$(version).zip --publish --output text

deploy/reducer: package/reducer
	aws lambda update-function-code --function-name location-reducer --zip-file fileb://bin/reducer-$(version).zip  --publish --output text

deploy/static: ## Deploys index.html to S3 and initiates CloudFront invalidation.
	@echo "Deploying static site version $(version)"
	$(shell sed -i -e 's,<version>.*</version>,<version>$(version)</version>,' ./static/index.html)
	aws s3 cp ./static/index.html s3://afrench-locations/index.html
	aws cloudfront create-invalidation --distribution-id E3BO48FU8DO9Z6 --paths /index.html

.DEFAULT_GOAL := help
.PHONY: help
help: ## Print Makefile help text.
	@grep -E '^[a-zA-Z_\/-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| awk 'BEGIN {FS = ":.*?## "}; \
		{printf "\033[36m%-20s\033[0m%s\n", $$1, $$2}'
