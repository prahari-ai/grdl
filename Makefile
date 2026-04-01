.PHONY: build test bench clean docker

BINARY := prahari
VERSION := 0.1.0
LDFLAGS := -s -w

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/prahari

test:
	go test -v ./tests/

bench:
	go test -bench=. -benchmem ./tests/

clean:
	rm -f $(BINARY)
	rm -rf compiled/

docker:
	docker build -t prahari-cfais:$(VERSION) -f core/sandbox/Dockerfile .

# Quick smoke test: compile + evaluate
smoke: build
	./$(BINARY) validate examples/templates/enterprise-agent-governance.grdl.yaml
	./$(BINARY) validate examples/templates/dao-governance.grdl.yaml
	./$(BINARY) validate examples/templates/ai-safety.grdl.yaml
	./$(BINARY) compile examples/templates/enterprise-agent-governance.grdl.yaml
	./$(BINARY) evaluate examples/templates/enterprise-agent-governance.grdl.yaml examples/test-action-denied.json
	./$(BINARY) evaluate examples/templates/enterprise-agent-governance.grdl.yaml examples/test-action-allowed.json
	@echo "smoke tests passed"
