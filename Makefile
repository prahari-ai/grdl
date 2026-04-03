.PHONY: build test bench clean docker smoke

BINARY := prahari
VERSION := 0.2.0
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

smoke: build
	@echo "=== Validate ==="
	./$(BINARY) validate examples/templates/enterprise-agent-governance.grdl.yaml
	./$(BINARY) validate examples/templates/dao-governance.grdl.yaml
	./$(BINARY) validate examples/templates/ai-safety.grdl.yaml
	@echo "\n=== Compile (3 backends) ==="
	./$(BINARY) compile examples/templates/enterprise-agent-governance.grdl.yaml --backend openshell --output-dir compiled/openshell
	./$(BINARY) compile examples/templates/enterprise-agent-governance.grdl.yaml --backend docker --output-dir compiled/docker
	./$(BINARY) compile examples/templates/enterprise-agent-governance.grdl.yaml --backend standalone --output-dir compiled/standalone
	@echo "\n=== Evaluate ==="
	./$(BINARY) evaluate examples/templates/enterprise-agent-governance.grdl.yaml examples/test-action-denied.json
	./$(BINARY) evaluate examples/templates/enterprise-agent-governance.grdl.yaml examples/test-action-allowed.json
	@echo "\nsmoke tests passed"
