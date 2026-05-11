.PHONY: build test lint tidy clean

BINARY := ../bin/luna-interceptor
GO     := go

build:
	@mkdir -p ../bin
	$(GO) build -ldflags="-s -w" -o $(BINARY) ./main.go
	@echo "Built: $(BINARY)"

test:
	$(GO) test ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found — install from https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run ./...

tidy:
	$(GO) mod tidy

clean:
	rm -f $(BINARY)

run: build
	SYSADM_HOSTS_FILE=hosts.yaml $(BINARY)
