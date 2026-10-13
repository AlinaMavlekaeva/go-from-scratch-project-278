build:
	go build -o bin/app ./cmd/api
lint:
	golangci-lint run
lint-fix:
	golangci-lint run --fix
tests:
	go test code -cover