include .env
export $(shell sed 's/=.*//' .env)

VERSION := "-X main.Version=$(shell git rev-parse --short HEAD)_$(shell date -u +%Y-%m-%d_%H:%M:%S)"

default: run

init:
	go install github.com/go-swagger/go-swagger/cmd/swagger@latest

run:
	@go run -ldflags $(VERSION) $(RACE_RUN_FLAG) ./cmd/main/main.go

bp:
	@env GOOS=linux GOARCH=386 go build -o goserver -ldflags $(VERSION) ./cmd/main/main.go

b: gen lint test bp

test:
	@go test ./...

lint:
	@golangci-lint run
	@echo Linters are ok!

gen:
	@echo "Removing generated files..."
	@rm -rf ./generated
	@mkdir -p generated
	@swagger generate model -f ./swagger.yml -t ./generated --accept-definitions-only
	@go generate ./...
