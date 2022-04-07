.DEFAULT_GOAL 	:= help

compile: ## compile:
	@mkdir -p build
	@go build -o build/posgonitor cmd/posgonitor/main.go

run: ## run
	@./build/posgonitor

all: compile run ## build and run

test:
	@go test ./...

test-cover: ## tests with coverage
	@mkdir -p coverage
	@go test -coverprofile=coverage/coverage.out -covermode=count ./...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html

gomod_tidy:
	@go mod tidy

gofmt:
	@go fmt -x ./...

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'