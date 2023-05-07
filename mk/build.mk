# --------------------------------------------------
# Build tooling
# --------------------------------------------------

EXE_NAME := $(APP_NAME)
VERSION ?= main
COMMIT ?= $(shell git rev-parse --short HEAD)
OS_ARCH ?= $(shell go version | awk '{print $$4;}')
GO_VERSION ?= $(shell go version | awk '{print $$3;}')
BUILD_DATE ?= $(shell date $(APP_DATE_FORMAT))
PACKAGE := github.com/Kong/koko-slack-bot/internal/
define LDFLAGS
-X $(PACKAGE).version=$(VERSION) \
-X $(PACKAGE).commit=$(COMMIT) \
-X $(PACKAGE).osArch=$(OS_ARCH) \
-X $(PACKAGE).goVersion=$(GO_VERSION) \
-X $(PACKAGE).buildDate=$(BUILD_DATE)
endef

.PHONY: build
build:
	@CGO_ENABLED=0 go build \
		-ldflags "$(LDFLAGS)" \
		-o bin/$(EXE_NAME) \
		main.go

.PHONY:
lint: install-tools
	@golangci-lint run -v ./...