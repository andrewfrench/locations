version := $(shell date -u +v%Y.%m.%d.%H%M%S)
bootstrap:
	go run cmd/bootstrap/bootstrap.go


build:
	@echo "Building $(version)"
	GOOS=linux go build -o bin/receiver/handler  cmd/receiver/receiver.go
	GOOS=linux go build -o bin/reducer/handler   cmd/reducer/reducer.go

package:
	@echo "Packaging $(version)"
	zip -jo bin/receiver-$(version).zip  bin/receiver/handler
	zip -jo bin/reducer-$(version).zip   bin/reducer/handler

deploy:
	@echo "Deploying $(version)"
	aws lambda update-function-code --function-name location-receiver  --zip-file fileb://bin/receiver-$(version).zip --publish
	aws lambda update-function-code --function-name location-reducer   --zip-file fileb://bin/reducer-$(version).zip  --publish


