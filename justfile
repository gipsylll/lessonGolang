# Версия по умолчанию
default_version := "0.1.0"

# Параметры сборки
binary   := "api"
image    := "sushkov-api"
main_pkg := "./cmd/api"

# ── Сборка ───────────────────────────────────────────────────────────────────

build version=default_version:
    #!/usr/bin/env sh
    commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    built=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    echo "Building {{binary}} v{{version}}"
    go build \
      -ldflags "-s -w -X main.Version={{version}} -X main.GitCommit=$commit -X main.BuildTime=$built" \
      -o {{binary}} {{main_pkg}}

release version:
    #!/usr/bin/env sh
    echo "Creating release v{{version}}"
    git tag -a "v{{version}}" -m "Release v{{version}}"
    git push origin "v{{version}}"
    commit=$(git rev-parse --short HEAD)
    built=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    go build \
      -ldflags "-s -w -X main.Version={{version}} -X main.GitCommit=$commit -X main.BuildTime=$built" \
      -o {{binary}} {{main_pkg}}
    echo "Release v{{version}} created!"

release-all version:
    #!/usr/bin/env sh
    commit=$(git rev-parse --short HEAD)
    built=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    flags="-s -w -X main.Version={{version}} -X main.GitCommit=$commit -X main.BuildTime=$built"
    mkdir -p dist
    GOOS=linux   GOARCH=amd64 go build -ldflags "$flags" -o dist/{{binary}}-linux-amd64       {{main_pkg}}
    GOOS=darwin  GOARCH=amd64 go build -ldflags "$flags" -o dist/{{binary}}-darwin-amd64      {{main_pkg}}
    GOOS=windows GOARCH=amd64 go build -ldflags "$flags" -o dist/{{binary}}-windows-amd64.exe {{main_pkg}}
    echo "Release artifacts in dist/"

# ── Разработка ───────────────────────────────────────────────────────────────

# Запуск без hot-reload (загружает .env, требует запущенной инфры: just infra)
run:
    #!/usr/bin/env sh
    set -a && . ./.env && set +a
    go run {{main_pkg}}

# Hot-reload (требует: go install github.com/air-verse/air@latest)
dev:
    #!/usr/bin/env sh
    set -a && . ./.env && set +a
    air --build.cmd "go build -o {{binary}} {{main_pkg}}" --build.bin "./{{binary}}"

# Применить миграции вручную (требует запущенной инфры: just infra)
migrate:
    #!/usr/bin/env sh
    set -a && . ./.env && set +a
    GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" \
    goose -dir migrations up

# Откатить последнюю миграцию
migrate-down:
    #!/usr/bin/env sh
    set -a && . ./.env && set +a
    GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" \
    goose -dir migrations down

# Только postgres + redis для локальной разработки
infra:
    docker compose up -d postgres redis

infra-down:
    docker compose down

# ── Docker ───────────────────────────────────────────────────────────────────

docker-build version=default_version:
    docker build -t {{image}}:{{version}} --build-arg VERSION={{version}} .

docker-push version:
    docker push {{image}}:{{version}}

up version=default_version:
    VERSION={{version}} docker compose up -d --build

down:
    docker compose down

logs:
    docker compose logs -f api

# ── Качество кода ─────────────────────────────────────────────────────────────

lint:
    golangci-lint run ./...

lint-fix:
    golangci-lint run --fix ./...

test:
    go test ./... -v -race

test-cover:
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out

# ── Зависимости ──────────────────────────────────────────────────────────────

deps:
    go mod download
    go mod tidy

# ── CI ───────────────────────────────────────────────────────────────────────

ci: lint test build

# ── Помощь ───────────────────────────────────────────────────────────────────

help:
    @just --list
