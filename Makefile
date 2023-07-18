build:
	@go build -o bin/aadvcs

run:
	@./bin/aadvcs
	
test:
	go test -v ./...