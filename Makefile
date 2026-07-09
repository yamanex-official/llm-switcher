BINARY := llm-switcher
SRC := ./cmd/llm-switcher

.PHONY: build build-all vet tidy clean

build:
	go build -o bin/$(BINARY) $(SRC)

build-all:
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY)-windows-amd64.exe $(SRC)
	GOOS=darwin  GOARCH=arm64 go build -o bin/$(BINARY)-darwin-arm64 $(SRC)
	GOOS=linux   GOARCH=amd64 go build -o bin/$(BINARY)-linux-amd64 $(SRC)

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin
