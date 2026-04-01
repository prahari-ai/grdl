FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /prahari ./cmd/prahari

# ---

FROM alpine:3.20

LABEL maintainer="Jagan Mohan Kataru <klk@pramaan.work>"
LABEL description="Prahari CFAIS Governance Policy Engine for OpenShell"
LABEL version="0.1.0"

RUN apk add --no-cache curl \
 && addgroup -S sandbox && adduser -S sandbox -G sandbox -h /sandbox

COPY --from=builder /prahari /usr/local/bin/prahari

# Copy default ruleset (enterprise template — swap for dao or ai-safety)
COPY examples/templates/enterprise-agent-governance.grdl.yaml /sandbox/cfais-rules/default.grdl.yaml

RUN mkdir -p /sandbox/cfais-audit /sandbox/cfais-engine \
 && chown -R sandbox:sandbox /sandbox

USER sandbox
WORKDIR /sandbox

EXPOSE 9700

HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD curl -f http://127.0.0.1:9700/healthz || exit 1

ENTRYPOINT ["prahari", "serve", "/sandbox/cfais-rules/default.grdl.yaml", "--addr", ":9700"]
