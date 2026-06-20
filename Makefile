deps:
	go mod download
	go mod tidy

swagger:
	~/go/bin/swag init -g cmd/api/main.go

swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

clean:
	rm -rf bin/