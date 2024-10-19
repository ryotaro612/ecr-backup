PROJECT_ROOT := $(dir $(lastword $(MAKEFILE_LIST)))
.PHONY: help test all

help: ## display this message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

DIST_DIR := $(PROJECT_ROOT)dist
BIN := $(DIST_DIR)/ecrbk
all: $(BIN) ## build the artifacts

test: ## Run the tests
	go test -tags test -v ./...


clean: ## delete the intermediate files
	rm -rf $(DIST_DIR)

$(BIN): $(shell find . -type f -regex "[^#]+\.go")
	mkdir -p $(DIST_DIR)
	go build -o $(BIN) ./cmd/ecrbk.go


