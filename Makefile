test:
	go test ./... -v

fmt-check:
	@unformatted="$$(gofmt -l ./internal ./cmd)"; \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed on:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

check: fmt-check
	go vet ./...
	go run ./cmd/lexcheck
	go test ./... -v