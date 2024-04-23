# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

export PATH := $(abspath bin/):${PATH}
OS = $(shell uname | tr A-Z a-z)

.PHONY: ci
ci: build test lint ## Run all builds and checks

.PHONY: build
build: ## Build all binaries
	@mkdir -p build
	go build -trimpath -o build/app .

.PHONY: run
run: build ## Build and run the application
	build/app

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: lint
lint: ## Run linter
	golangci-lint run

# Dependency versions
GOLANGCI_VERSION ?= 1.57.2
DAGGER_VERSION ?= 0.11.1

deps: bin/golangci-lint

bin/golangci-lint:
	@mkdir -p bin
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}

bin/dagger:
	@mkdir -p bin
	curl -L https://dl.dagger.io/dagger/install.sh | sh
	@echo ${HELLO}

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'







































HELLO := "🦄 🌈 🦄 🌈 🦄 🌈"
