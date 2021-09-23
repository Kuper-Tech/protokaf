include scripts/*.mk

APP_NAME=protokaf

run:	## Run protokaf
	@go run .

test:	## Run tests
	go test -cover -p 1 -count=1 ./...

none:
	sleep 31536000

lint:	## Run linter
	golangci-lint run

install:
	go install .
