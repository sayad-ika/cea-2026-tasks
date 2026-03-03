run:
	go run ./cmd/main.go

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap ./cmd/main.go
	zip function.zip bootstrap

test:
	go test ./...
