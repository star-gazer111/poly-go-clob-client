.PHONY: test race lint vuln sec tidy

test:
	go test ./... -count=1

race:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

vuln:
	govulncheck ./...

sec:
	gosec ./...

tidy:
	go mod tidy
