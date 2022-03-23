lint:
	golangci-lint run -v ./...

test:
	go test -count=1 -v ./...

run:
	go run cmd/todo/main.go