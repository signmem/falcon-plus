FILES = $(shell ls -1 *.go | grep -v _test.go)

test:
	@go test -v ./...

fmt:
	@go fmt ./...

dev:
	@go run ${FILES} -disableSend -enableCnt

send:
	@go run ${FILES}

race:
	@go run -race ${FILES} -disableSend

build:
	@GOOS=linux GOARCH=amd64 go build -o build/falcon-collectord

.PHONY: build
