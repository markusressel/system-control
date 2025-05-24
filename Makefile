.PHONY: help test build run deploy clean

GO_FLAGS   ?=
NAME       := system-control
PACKAGE    := github.com/markusressel/$(NAME)
OUTPUT_DIR  := "bin/"
GIT_REV    ?= $(shell git rev-parse --short HEAD)
SOURCE_DATE_EPOCH ?= $(shell date +%s)
DATE       ?= $(shell date -u -d @${SOURCE_DATE_EPOCH} +"%Y-%m-%dT%H:%M:%SZ")
VERSION    ?= 0.8.1

test:   ## Run all tests
	@go clean --testcache && go test -v ./...

build: clean
	go build -o ${OUTPUT_DIR}${NAME} main.go

run: build
	./${OUTPUT_DIR}${NAME}

deploy-custom: build
	mkdir -p ~/.custom/bin/
	cp ./${OUTPUT_DIR}${NAME} ~/.custom/bin/${NAME}
	system-control completion fish > ~/.config/fish/completions/system-control.fish


deploy: build
	sudo cp ./${OUTPUT_DIR}${NAME} /usr/bin/${NAME}
	system-control completion fish > ~/.config/fish/completions/system-control.fish

clean:
	go clean
	rm -rf ${OUTPUT_DIR}${NAME}