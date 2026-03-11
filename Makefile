.PHONY: test race lint vuln sec tidy

test:
	go test ./... -count=1

race:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

security:
	govulncheck ./...
	gosec ./...

check: fmt tidy lint test

