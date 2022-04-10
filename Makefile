.DEFAULT_GOAL 	:= help

compile: ## compile:
	@mkdir -p build
	@go build -o build/posgonitor cmd/main.go

run: ## run
	@./build/posgonitor --config=config.yaml

all: compile run ## build and run

test:
	@go test -timeout 30s ./...

integration-tests:
	@go test -timeout 7m -tags=integration ./...

test-cover: ## tests with coverage
	@mkdir -p coverage
	@go test -coverprofile=coverage/coverage.out -covermode=count ./...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html

install-deps: ## Install some project dependencies
	@git clone https://github.com/stdevMac/courtney
	@(cd courtney && go get  ./... && go build courtney.go)
	@go get ./...

codecov-test:
	mkdir -p coverage
	courtney/courtney -v -o coverage/coverage.out ./...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html

gomod_tidy:
	@go mod tidy

gofmt:
	@go fmt -x ./...

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'