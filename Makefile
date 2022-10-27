GIT_TAG?= $(shell git describe --always --tags)
MODULE=github.com/ericogr/azenv
BIN = azenv
FMT_CMD = $(gofmt -s -l -w $(find . -type f -name '*.go' -not -path './vendor/*') | tee /dev/stderr)
IMAGE_REPO = ericogr
BUILD_DATE ?=  $(shell date "+%Y-%m-%d %H:%M:%S")
BUILDFLAGS := "-w -s -X '$(MODULE)/cmd.Version=$(GIT_TAG)' -X '$(MODULE)/cmd.GitTag=$(GIT_TAG)' -X '$(MODULE)/cmd.BuildDate=$(BUILD_DATE)'"
CGO_ENABLED = 0
GO := GO111MODULE=on go
GO_NOMOD :=GO111MODULE=off go
GOPATH ?= $(shell $(GO) env GOPATH)
GOBIN ?= $(GOPATH)/bin
GOLINT ?= $(GOBIN)/golint
GOSEC ?= $(GOBIN)/gosec
GO_MINOR_VERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
GOVULN_MIN_VERSION = 18
GO_VERSION = 1.18

default:
	$(MAKE) build

install-govulncheck:
	@if [ $(GO_MINOR_VERSION) -gt $(GOVULN_MIN_VERSION) ]; then \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi

fmt:
	@echo "FORMATTING"
	@FORMATTED=`$(GO) fmt ./...`
	@([ ! -z "$(FORMATTED)" ] && printf "Fixed unformatted files:\n$(FORMATTED)") || true

lint:
	@echo "LINTING: golint"
	$(GO_NOMOD) get -u golang.org/x/lint/golint
	$(GOLINT) -set_exit_status ./...
	@echo "VETTING"
	$(GO) vet ./...

golangci:
	@echo "LINTING: golangci-lint"
	golangci-lint run

govulncheck: install-govulncheck
	@echo "CHECKING VULNERABILITIES"
	@if [ $(GO_MINOR_VERSION) -gt $(GOVULN_MIN_VERSION) ]; then \
		govulncheck ./...; \
	fi

build:
	go build -o $(BIN) .

clean:
	rm -rf build vendor dist
	rm -f release image $(BIN)

release:
	@echo "Releasing the $(BIN) binary..."
	goreleaser release

build-linux:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -ldflags=$(BUILDFLAGS) -o $(BIN) .

image:
	@echo "Building the Docker image..."
	docker build -t $(IMAGE_REPO)/$(BIN):$(GIT_TAG) --build-arg GO_VERSION=$(GO_VERSION) .
	docker tag $(IMAGE_REPO)/$(BIN):$(GIT_TAG) $(IMAGE_REPO)/$(BIN):latest
	touch image

image-push: image
	@echo "Pushing the Docker image..."
	docker push $(IMAGE_REPO)/$(BIN):$(GIT_TAG)
	docker push $(IMAGE_REPO)/$(BIN):latest

.PHONY: build clean release image image-push