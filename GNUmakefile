BINNARY_NAME="terraform-provider-cop"

# Default target
default: build

.PHONY: build
build:
	go build -o $(BINNARY_NAME)

.PHONY: clean
clean:
	rm -f $(BINNARY_NAME)

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

# Run linting with golangci-lint
.PHONY: lint
lint:
	golangci-lint run ./...
