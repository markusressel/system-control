.PHONY: help test build run deploy clean

BINARY_NAME = system-control
OUTPUT_DIR  = bin/
GO_FLAGS   ?=
NAME       := system-control
PACKAGE    := github.com/markusressel/$(NAME)
GIT_REV    ?= $(shell git rev-parse --short HEAD)
SOURCE_DATE_EPOCH ?= $(shell date +%s)
DATE       ?= $(shell date -u -d @${SOURCE_DATE_EPOCH} +"%Y-%m-%dT%H:%M:%SZ")
VERSION    ?= 0.8.1

test:   ## Run all tests
	@go clean --testcache && go test -v ./...

build: clean
	go build -o ${OUTPUT_DIR}${BINARY_NAME} main.go

run: build
	./${OUTPUT_DIR}${BINARY_NAME}

deploy-custom: build
	mkdir -p ~/.custom/bin/
	cp ./${OUTPUT_DIR}${BINARY_NAME} ~/.custom/bin/${BINARY_NAME}

deploy: build
	sudo cp ./${OUTPUT_DIR}${BINARY_NAME} /usr/bin/${BINARY_NAME}

clean:
	go clean
	rm -rf ${OUTPUT_DIR}${BINARY_NAME}