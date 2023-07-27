build:
	@go build -o bin/aadvcs

run:
	@./bin/aadvcs
	
test:
	go test -v ./...

.PHONY: build-cli
build-cli:
	go build -o aadvcs.exe main.go