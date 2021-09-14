BINARY_NAME=system-control
OUTPUT_DIR=bin/

build:
	go build -o ${OUTPUT_DIR}${BINARY_NAME} main.go

run:
	go build -o ${OUTPUT_DIR}${BINARY_NAME} main.go
	./${OUTPUT_DIR}${BINARY_NAME}

clean:
	go clean
	rm ${OUTPUT_DIR}${BINARY_NAME}