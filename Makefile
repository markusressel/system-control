BINARY_NAME=system-control
OUTPUT_DIR=bin/

build: clean
	go build -o ${OUTPUT_DIR}${BINARY_NAME} main.go

run: build
	./${OUTPUT_DIR}${BINARY_NAME}

deploy: build
	mkdir -p ~/.custom/bin/
	cp ./${OUTPUT_DIR}${BINARY_NAME} ~/.custom/bin/${BINARY_NAME}

clean:
	go clean
	rm -rf ${OUTPUT_DIR}${BINARY_NAME}