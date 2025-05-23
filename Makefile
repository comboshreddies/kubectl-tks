default: help

PROJECTNAME=$(shell basename "$(PWD)")

BIN_FOLDER=bin

BIN_FOLDER_MACOS=${BIN_FOLDER}/amd64/darwin
BIN_FOLDER_LINUX=${BIN_FOLDER}/amd64/linux
BIN_FOLDER_MACOS_ARM=${BIN_FOLDER}/arm64/darwin
BIN_FOLDER_LINUX_ARM=${BIN_FOLDER}/arm64/linux
BIN_FOLDER_SCRATCH=${BIN_FOLDER}/amd64/scratch
BIN_NAME=${PROJECTNAME}

MAKEFLAGS += --silent

LDFLAGS=

setup: mod-download

compile: clean generate fmt vet test build

clean:
	@echo "  >  Cleaning build cache"
	@-rm -rf ${BIN_FOLDER}/amd64 ${BIN_FOLDER}/arm64 ${BIN_FOLDER}/${BIN_NAME} \
		&& go clean ./...

build:
	@echo "  >  Building binary"
	@go build \
		-ldflags="${LDFLAGS}" \
		-o ${BIN_FOLDER}/${BIN_NAME} \
		"${CLI_MAIN_FOLDER}"

build-all: build-macos build-linux build-macos-arm build-linux-arm build-alpine-scratch

build-macos:
	@echo "  >  Building binary for MacOS"
	@GOOS=darwin GOARCH=amd64 \
		go build \
		-ldflags="${LDFLAGS}" \
		-o ${BIN_FOLDER_MACOS}/${BIN_NAME} \
		"${CLI_MAIN_FOLDER}"

build-linux:
	@echo "  >  Building binary for Linux"
	@GOOS=linux GOARCH=amd64 \
		go build \
		-ldflags="${LDFLAGS}" \
		-o ${BIN_FOLDER_LINUX}/${BIN_NAME} \
		"${CLI_MAIN_FOLDER}"

build-macos-arm:
	@echo "  >  Building binary for MacOS ARM"
	@GOOS=darwin GOARCH=arm64 \
		go build \
		-ldflags="${LDFLAGS}" \
		-o ${BIN_FOLDER_MACOS_ARM}/${BIN_NAME} \
		"${CLI_MAIN_FOLDER}"

build-linux-arm:
	@echo "  >  Building binary for Linux ARM"
	@GOOS=linux GOARCH=arm64 \
		go build \
		-ldflags="${LDFLAGS}" \
		-o ${BIN_FOLDER_LINUX_ARM}/${BIN_NAME} \
		"${CLI_MAIN_FOLDER}"


# Alpine & scratch base images use musl instead of gnu libc, thus we need to add additional parameters on the build
build-alpine-scratch:
	@echo "  >  Building binary for Alpine/Scratch"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build \
		-ldflags="${LDFLAGS}" \
		-a -installsuffix cgo \
		-o ${BIN_FOLDER_SCRATCH}/${BIN_NAME} \
		"${CLI_MAIN_FOLDER}"

fmt:
	@echo "  >  Formatting code"
	@go fmt ./...

generate:
	@echo "  >  Go generate"
	@if !type "stringer" > /dev/null 2>&1; then \
		go install golang.org/x/tools/cmd/stringer@latest; \
	fi
	@go generate ./...

mod-download:
	@echo "  >  Download dependencies..."
	@go mod download && go mod tidy

test:
	@echo "  >  Executing unit tests"
	@go test -v -timeout 60s -race ./...

test-colorized:
	@echo "  >  Executing unit tests"
	@if ! type "richgo" > /dev/null 2>&1; then \
		go install github.com/kyoh86/richgo@latest; \
	fi
	@richgo test -v -timeout 60s -race ./...

vet:
	@echo "  >  Checking code with vet"
	@go vet ./...

.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
