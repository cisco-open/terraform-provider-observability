TOP_LEVEL=$(shell git rev-parse --show-toplevel)
PLUGIN_NAME := terraform-provider-observability
TOOLS_DIR := $(TOP_LEVEL)/.tools
ADDLICENSE := $(TOOLS_DIR)/addlicense
ADDLICENSE_VERSION := 1.1.1
FSOC := $(TOOLS_DIR)/fsoc
FSOC_VERSION := 0.67.0
GOLANGCI_LINT := $(TOOLS_DIR)/golangci-lint
GOLANGCI_LINT_VERSION := 1.57.2

$(ADDLICENSE):
	mkdir -p $(TOOLS_DIR)
	wget https://github.com/google/addlicense/releases/download/v$(ADDLICENSE_VERSION)/addlicense_$(ADDLICENSE_VERSION)_Linux_x86_64.tar.gz
	tar -xvf addlicense_$(ADDLICENSE_VERSION)_Linux_x86_64.tar.gz -C $(TOOLS_DIR) addlicense
	rm addlicense_$(ADDLICENSE_VERSION)_Linux_x86_64.tar.gz

$(FSOC):
	mkdir -p $(TOOLS_DIR)
	wget https://github.com/cisco-open/fsoc/releases/download/v$(FSOC_VERSION)/fsoc-linux-amd64.tar.gz
	tar -xvf fsoc-linux-amd64.tar.gz -C $(TOOLS_DIR) fsoc-linux-amd64
	mv $(TOOLS_DIR)/fsoc-linux-amd64 $(TOOLS_DIR)/fsoc
	rm fsoc-linux-amd64.tar.gz

$(GOLANGCI_LINT):
	mkdir -p $(TOOLS_DIR)
	wget https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz
	tar --strip-components 1 -xvf golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz -C $(TOOLS_DIR) golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64/golangci-lint
	rm golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz

.PHONY: all
all: lint test check-license

.PHONY: check-license
check-license: $(ADDLICENSE)
	@echo "verifying license headers"
	$(TOOLS_DIR)/addlicense -check .

.PHONY: add-license
add-license: $(ADDLICENSE)
	@echo "adding license headers, please commit any modified files"
	$(ADDLICENSE) -s -v -l mpl .

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

.PHONY: lint-fix
lint-fix: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix ./...

.PHONY: test
test:
	TF_ACC=1 TF_ACC_PROVIDER_NAMESPACE=cisco-open  go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: plugin
plugin:
	go build -o $(PLUGIN_NAME)

.PHONY: clean
clean:
	rm -rf $(PLUGIN_NAME) $(TOOLS_DIR)

