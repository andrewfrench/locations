version := $(shell date -u +v%Y.%m.%d.%H%M%S)
bootstrap:
	go run cmd/bootstrap/bootstrap.go

build:
	@echo "Building $(version)"
	GOOS=linux go build -o bin/receiver/handler  cmd/receiver/receiver.go
	GOOS=linux go build -o bin/reducer/handler   cmd/reducer/reducer.go

package: build
	@echo "Packaging $(version)"
	zip -jo bin/receiver-$(version).zip  bin/receiver/handler
	zip -jo bin/reducer-$(version).zip   bin/reducer/handler

deploy: build package
	@echo "Deploying $(version)"
	aws lambda update-function-code --function-name location-receiver  --zip-file fileb://bin/receiver-$(version).zip --publish
	aws lambda update-function-code --function-name location-reducer   --zip-file fileb://bin/reducer-$(version).zip  --publish

deploy-static:
	@echo "Deploying static site version $(version)"
	$(shell sed -i -e 's,var version = ".*" // inserted by make,var version = "$(version)" // inserted by make,' ./static/index.html)
	# aws s3 cp ./static/index.html  s3://afrench-locations/index.html
	# aws cloudfront create-invalidation --distribution-id E3BO48FU8DO9Z6 --paths /index.html --paths /version.txt

