default: help

PROJECTNAME=$(shell basename "$(PWD)")

BIN_FOLDER=bin

AMD64=amd64
ARM64=arm64
MAC=darwin
LINUX=linux
BIN_FOLDER_AMD64=${BIN_FOLDER}/${AMD64}
BIN_FOLDER_ARM64=${BIN_FOLDER}/${ARM64}
BIN_NAME=${PROJECTNAME}
RELEASE_DIR=release
RELEASE_VERSION:=$$(${BIN_FOLDER}/${PROJECTNAME} version )

MAKEFLAGS += --silent

LDFLAGS=

setup: mod-download

compile: clean generate fmt vet test build

clean:
	@echo "  >  Cleaning build cache"
	@-rm -rf ${BIN_FOLDER}/amd64 ${BIN_FOLDER}/arm64 ${BIN_FOLDER}/${BIN_NAME} ${RELEASE_DIR} \
		&& go clean ./...

build:
	@echo "  >  Building binary"
	@go build \
		-ldflags="${LDFLAGS}" \
		-o ${BIN_FOLDER}/${BIN_NAME} 

build-all: build-macos build-linux build-macos-arm build-linux-arm

build-macos:
	@echo "  >  Building binary for MacOS"
	@GOOS=darwin GOARCH=amd64 \
		go build \
		-ldflags="${LDFLAGS}" \
		-o ${BIN_FOLDER_AMD64}/${MAC}/${BIN_NAME}

build-linux:
	@echo "  >  Building binary for Linux"
	@GOOS=linux GOARCH=amd64 \
		go build \
		-ldflags="${LDFLAGS}" \
		-o ${BIN_FOLDER_AMD64}/${LINUX}/${BIN_NAME}

build-macos-arm:
	@echo "  >  Building binary for MacOS ARM"
	@GOOS=darwin GOARCH=arm64 \
		go build \
		-ldflags="${LDFLAGS}" \
		-o ${BIN_FOLDER_ARM64}/${MAC}/${BIN_NAME}

build-linux-arm:
	@echo "  >  Building binary for Linux ARM"
	@GOOS=linux GOARCH=arm64 \
		go build \
		-ldflags="${LDFLAGS}" \
		-o ${BIN_FOLDER_ARM64}/${LINUX}/${BIN_NAME}

pack-macos: build build-macos
	@echo " > Packing release " ${RELEASE_VERSION} " " ${AMD64} " for MacOs" 
	@mkdir -p ${RELEASE_DIR}/${RELEASE_VERSION}/${AMD64}/${MAC}/
	@cp ${BIN_FOLDER_AMD64}/${MAC}/${BIN_NAME} ${RELEASE_DIR}/${RELEASE_VERSION}/${AMD64}/${MAC}/
	@cp sequences.json ${RELEASE_DIR}/${RELEASE_VERSION}/${AMD64}/${MAC}/ 
	@(cd ${RELEASE_DIR}/${RELEASE_VERSION}/${AMD64}/${MAC}/ ; tar czvf ../../${BIN_NAME}.${AMD64}_${MAC}.tgz ${BIN_NAME} sequences.json)

pack-macos-arm: build build-macos-arm
	@echo " > Packing release " ${RELEASE_VERSION} " " ${ARM64} " for MacOs" 
	@mkdir -p ${RELEASE_DIR}/${RELEASE_VERSION}/${ARM64}/${MAC}/
	@cp ${BIN_FOLDER_ARM64}/${MAC}/${BIN_NAME} ${RELEASE_DIR}/${RELEASE_VERSION}/${ARM64}/${MAC}/
	@cp sequences.json ${RELEASE_DIR}/${RELEASE_VERSION}/${ARM64}/${MAC}/ 
	@(cd ${RELEASE_DIR}/${RELEASE_VERSION}/${ARM64}/${MAC}/ ; tar czvf ../../${BIN_NAME}.${ARM64}_${MAC}.tgz ${BIN_NAME} sequences.json)

pack-linux: build build-linux
	@echo " > Packing release " ${RELEASE_VERSION} " " ${AMD64} " for Linux" 
	@mkdir -p ${RELEASE_DIR}/${RELEASE_VERSION}/${AMD64}/${LINUX}/
	@cp ${BIN_FOLDER_AMD64}/${LINUX}/${BIN_NAME} ${RELEASE_DIR}/${RELEASE_VERSION}/${AMD64}/${LINUX}/
	@cp sequences.json ${RELEASE_DIR}/${RELEASE_VERSION}/${AMD64}/${LINUX}/ 
	@(cd ${RELEASE_DIR}/${RELEASE_VERSION}/${AMD64}/${LINUX}/ ; tar czvf ../../${BIN_NAME}.${AMD64}_${LINUX}.tgz ${BIN_NAME} sequences.json)

pack-linux-arm: build build-linux-arm
	@echo " > Packing release " ${RELEASE_VERSION} " " ${ARM64} " for Linux" 
	@mkdir -p ${RELEASE_DIR}/${RELEASE_VERSION}/${ARM64}/${LINUX}/
	@cp ${BIN_FOLDER_ARM64}/${LINUX}/${BIN_NAME} ${RELEASE_DIR}/${RELEASE_VERSION}/${ARM64}/${LINUX}/
	@cp sequences.json ${RELEASE_DIR}/${RELEASE_VERSION}/${ARM64}/${LINUX}/ 
	@(cd ${RELEASE_DIR}/${RELEASE_VERSION}/${ARM64}/${LINUX}/ ; tar czvf ../../${BIN_NAME}.${ARM64}_${LINUX}.tgz ${BIN_NAME} sequences.json)

pack-all: pack-macos pack-macos-arm pack-linux pack-linux-arm

# Alpine & scratch base images use musl instead of gnu libc, thus we need to add additional parameters on the build
build-alpine-scratch:
	@echo "  >  Building binary for Alpine/Scratch"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build \
		-ldflags="${LDFLAGS}" \
		-a -installsuffix cgo \
		-o ${BIN_FOLDER_AMD64}/${BIN_NAME}_alpine \
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
