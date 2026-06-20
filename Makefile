deps:
	go mod download
	go mod tidy

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

clean:
	rm -rf bin/