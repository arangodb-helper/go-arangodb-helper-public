ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
REPO_NAME := go-helper

GO_VERSION ?= 1.20.5
GO_IMAGE ?= golang:$(GO_VERSION)

DOCKER_NAMESPACE ?= arangodb

SOURCES := $(shell find $(ROOT_DIR) -name '*.go')

TEST_IMAGE_NAME ?= $(DOCKER_NAMESPACE)/$(REPO_NAME)
TEST_IMAGE_TAG ?= latest
TEST_IMAGE ?= $(TEST_IMAGE_NAME):$(TEST_IMAGE_TAG)


.PHONY: run-unit-tests
run-unit-tests:
ifeq ("$(TEST_DEBUG)", "true")
	docker build --build-arg GO_VERSION=$(GO_VERSION) --build-arg TEST_DEBUG_PACKAGE=${TEST_DEBUG_PACKAGE} -f Dockerfile.unittest.debug -t $(TEST_IMAGE) .
	docker run -p 2345:2345 --cap-add=SYS_PTRACE --security-opt=seccomp:unconfined "${TEST_IMAGE}" ${TEST_OPTIONS}
else
	docker run --rm -v "${ROOT_DIR}":/usr/code -e CGO_ENABLED=0 -w /usr/code/ $(GO_IMAGE) go test $(TEST_OPTIONS) ./pkg/...
endif


.PHONY: tools
tools:
	@-mkdir -p bin
	@echo ">> Fetching golangci-lint linter"
	@GOBIN=$(ROOT_DIR)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3
	@echo ">> Fetching goimports"
	@GOBIN=$(ROOT_DIR)/bin go install golang.org/x/tools/cmd/goimports@v0.1.12
	@echo ">> Fetching license check"
	@GOBIN=$(ROOT_DIR)/bin go install github.com/google/addlicense@v1.0.0
	@echo ">> Fetching govulncheck"
	@GOBIN=$(ROOT_DIR)/bin go install golang.org/x/vuln/cmd/govulncheck@v0.1.0

.PHONY: linter
linter:
	@GO111MODULE=off $(ROOT_DIR)/bin/golangci-lint run ./...


.PHONY: fmt
fmt:
	@echo ">> Verify files style"
	@if [ X"$$($(ROOT_DIR)/bin/goimports -l $(SOURCES) | wc -l)" != X"0" ]; then echo ">> Style errors"; $(ROOT_DIR)/bin/goimports -l $(SOURCES); exit 1; fi


.PHONY: license
license:
	@echo ">> Verify license of files"
	@$(ROOT_DIR)/bin/addlicense -f "$(ROOT_DIR)/HEADER" -check $(SOURCES)


.PHONY: vulncheck
vulncheck:
	$(ROOT_DIR)/bin/govulncheck ./...


.PHONY: check
check: fmt license linter vulncheck
	echo "Verify project"
