## Default to the repo name if empty
ifndef BINARY_NAME
	override BINARY_NAME=app
endif

## Define the binary name
ifdef CUSTOM_BINARY_NAME
	override BINARY_NAME=$(CUSTOM_BINARY_NAME)
endif

## Set the binary release names
DARWIN=$(BINARY_NAME)-darwin
LINUX=$(BINARY_NAME)-linux
WINDOWS=$(BINARY_NAME)-windows.exe

.PHONY: test lint vet install

bench:  ## Run all benchmarks in the Go application
	@go test -bench=. -benchmem

build-go:  ## Build the Go application (locally)
	@go build -o bin/$(BINARY_NAME)

clean-mods: ## Remove all the Go mod cache
	@go clean -modcache

coverage: ## Shows the test coverage
	@go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

godocs: ## Sync the latest tag with GoDocs
	@test $(GIT_DOMAIN)
	@test $(REPO_OWNER)
	@test $(REPO_NAME)
	@test $(VERSION_SHORT)
	@curl https://proxy.golang.org/$(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)/@v/$(VERSION_SHORT).info

install: ## Install the application
	@go build -o $$GOPATH/bin/$(BINARY_NAME)

install-go: ## Install the application (Using Native Go)
	@go install $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)

lint: ## Run the golangci-lint application (install if not found)
	@#Travis (has sudo)
	@if [ "$(shell command -v golangci-lint)" = "" ] && [ $(TRAVIS) ]; then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.33.0 && sudo cp ./bin/golangci-lint $(go env GOPATH)/bin/; fi;
	@#AWS CodePipeline
	@if [ "$(shell command -v golangci-lint)" = "" ] && [ "$(CODEBUILD_BUILD_ID)" != "" ]; then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.33.0; fi;
	@#Github Actions
	@if [ "$(shell command -v golangci-lint)" = "" ] && [ "$(GITHUB_WORKFLOW)" != "" ]; then curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b $(go env GOPATH)/bin v1.33.0; fi;
	@#Brew - MacOS
	@if [ "$(shell command -v golangci-lint)" = "" ] && [ "$(shell command -v brew)" != "" ]; then brew install golangci-lint; fi;
	@echo "running golangci-lint..."
	@golangci-lint run

test: ## Runs vet, lint and ALL tests
	@$(MAKE) lint
	@echo "running tests..."
	@go test ./... -v

test-short: ## Runs vet, lint and tests (excludes integration tests)
	@$(MAKE) lint
	@echo "running tests (short)..."
	@go test ./... -v -test.short

test-ci: ## Runs all tests via CI (exports coverage)
	@$(MAKE) lint
	@echo "running tests (CI)..."
	@go test ./... -race -coverprofile=coverage.txt -covermode=atomic

test-ci-no-race: ## Runs all tests via CI (no race) (exports coverage)
	@$(MAKE) lint
	@echo "running tests (CI - no race)..."
	@go test ./... -coverprofile=coverage.txt -covermode=atomic

test-ci-short: ## Runs unit tests via CI (exports coverage)
	@$(MAKE) lint
	@echo "running tests (CI - unit tests only)..."
	@go test ./... -test.short -race -coverprofile=coverage.txt -covermode=atomic

uninstall: ## Uninstall the application (and remove files)
	@test $(BINARY_NAME)
	@test $(GIT_DOMAIN)
	@test $(REPO_OWNER)
	@test $(REPO_NAME)
	@go clean -i $(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)
	@rm -rf $$GOPATH/src/$(GIT_DOMAIN)/$(REPO_OWNER)/$(REPO_NAME)
	@rm -rf $$GOPATH/bin/$(BINARY_NAME)

update:  ## Update all project dependencies
	@go get -u ./... && go mod tidy

update-linter: ## Update the golangci-lint package (macOS only)
	@brew upgrade golangci-lint

vet: ## Run the Go vet application
	@echo "running go vet..."
	@go vet -v ./...