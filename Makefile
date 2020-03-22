tidy:
	go mod tidy
test:
	go test ./...
build:
	go build -o dist/slipee slipee.go
