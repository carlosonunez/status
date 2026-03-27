set shell         := ["bash", "-euo", "pipefail", "-c"]
set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

# Detect host platform so Docker-based builds cross-compile to the right target.
goos    := if os() == "macos" { "darwin" } else if os() == "windows" { "windows" } else { "linux" }
goarch  := if arch() == "aarch64" { "arm64" } else { "amd64" }
bin_ext := if os() == "windows" { ".exe" } else { "" }

# Show available recipes
default:
    @just --list

# ── build ─────────────────────────────────────────────────────────────────────

# Compile the status binary for the host platform (output: ./bin/status[.exe])
build:
    docker compose run --rm -e GOOS={{goos}} -e GOARCH={{goarch}} -e BIN_EXT={{bin_ext}} build

# Compile the status binary natively without Docker (output: ./bin/status[.exe])
build-local:
    go build -o bin/status{{bin_ext}} ./cmd/status

# ── test ──────────────────────────────────────────────────────────────────────

# Run all tests (unit + integration). Starts DynamoDB Local automatically.
test: test-unit test-integration

# Run unit tests only (no external services required)
test-unit:
    docker compose run --rm test-unit

# Run integration tests against DynamoDB Local
test-integration:
    docker compose run --rm test go test -tags integration ./... -v -count=1

# ── lint ──────────────────────────────────────────────────────────────────────

# Run golangci-lint
lint:
    docker compose run --rm lint

# ── run ───────────────────────────────────────────────────────────────────────

# Start the daemon with live reload (foreground)
run:
    docker compose up dev

# Start the daemon in the background
run-detached:
    docker compose up -d dev

# Tail daemon logs
logs:
    docker compose logs -f dev

# Stop all services
stop:
    docker compose down

# ── infrastructure ────────────────────────────────────────────────────────────

# Start DynamoDB Local in the background
dynamo-up:
    docker compose up -d dynamodb-local

# Stop DynamoDB Local
dynamo-down:
    docker compose stop dynamodb-local

# Create the token table in DynamoDB Local (idempotent)
dynamo-init:
    -docker compose run --rm awscli dynamodb create-table \
        --table-name status-tokens \
        --attribute-definitions \
            AttributeName=PK,AttributeType=S \
            AttributeName=SK,AttributeType=S \
        --key-schema \
            AttributeName=PK,KeyType=HASH \
            AttributeName=SK,KeyType=RANGE \
        --billing-mode PAY_PER_REQUEST \
        --endpoint-url http://dynamodb-local:8000 \
        --region us-east-1

# ── deploy ────────────────────────────────────────────────────────────────────

# Build a release binary and push to ECR (requires AWS credentials in config.yaml)
[unix]
deploy:
    sops exec-env config.yaml 'docker compose run --rm deploy'

# ── housekeeping ──────────────────────────────────────────────────────────────

# Remove build artifacts and stopped containers
[unix]
clean:
    docker compose down --volumes --remove-orphans
    rm -rf bin/

[windows]
clean:
    docker compose down --volumes --remove-orphans
    if (Test-Path bin) { Remove-Item -Recurse -Force bin }
