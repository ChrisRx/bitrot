BIN_DIR ?= bin

LDFLAGS := -s -w
GOFLAGS = -gcflags "all=-trimpath=$(PWD)" -asmflags "all=-trimpath=$(PWD)"

GO_BUILD_ENV_VARS := GO111MODULE=on CGO_ENABLED=0

##@ Building

.PHONY: bitrot

bitrot: ## Build the bitrot binary
	@$(GO_BUILD_ENV_VARS) go build -o $(BIN_DIR)/bitrot $(GOFLAGS) -ldflags '$(LDFLAGS)' ./cmd/bitrot
