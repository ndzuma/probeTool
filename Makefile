# Makefile for probeTool

BINARY_NAME=probe
GOBIN=$(shell go env GOPATH)/bin
PROBES_DIR=$(shell dirname $(GOBIN))/probes

.PHONY: probe
probe:
	go build -o $(BINARY_NAME) ./cmd/probe

.PHONY: install
install:
	go install ./cmd/probe

.PHONY: run
run: probe
	./$(BINARY_NAME)

.PHONY: clean
clean:
	rm -f $(BINARY_NAME) probes/*.md
