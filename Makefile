.PHONY: help test build run deploy clean

GO_FLAGS   ?=
NAME       := system-control
PACKAGE    := github.com/markusressel/$(NAME)
OUTPUT_DIR  := bin/
GIT_REV    ?= $(shell git rev-parse --short HEAD)
SOURCE_DATE_EPOCH ?= $(shell date +%s)
DATE       ?= $(shell date -u -d @${SOURCE_DATE_EPOCH} +"%Y-%m-%dT%H:%M:%SZ")
VERSION    ?= 0.8.1

test:   ## Run all tests
	@go clean --testcache && go test -v ./...

build: clean
	go build -o "${OUTPUT_DIR}${NAME}" main.go
	gcc -O3 cmd/sysctrl/main.c -o "${OUTPUT_DIR}sysctrl"

run: build
	"./${OUTPUT_DIR}${NAME}"

deploy-custom: build
	mkdir -p "~/.custom/bin/"
	cp "./${OUTPUT_DIR}${NAME}" "~/.custom/bin/${NAME}"
	system-control completion fish > ~/.config/fish/completions/system-control.fish


deploy: build
	-systemctl --user stop system-control.service 2>/dev/null || true
	sudo install -Dm755 "./${OUTPUT_DIR}${NAME}" "/usr/bin/${NAME}"
	sudo install -Dm755 "./${OUTPUT_DIR}sysctrl" "/usr/bin/sysctrl"
	system-control completion fish > ~/.config/fish/completions/system-control.fish
	$(MAKE) deploy-systemd

deploy-systemd:
	mkdir -p ~/.config/systemd/user/
	cp systemd/system-control.service ~/.config/systemd/user/
	systemctl --user daemon-reload
	systemctl --user enable --now system-control.service

clean:
	go clean
	rm -rf "${OUTPUT_DIR}${NAME}"