build:
	go build -o bin/app .
lint:
	golangci-lint run
lint-fix:
	golangci-lint run --fix
tests:
	go test . -cover