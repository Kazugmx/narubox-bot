#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

export DATABASE_URL="${DATABASE_URL:-postgres://narubox:narubox_dev@127.0.0.1:5432/narubox_dev?sslmode=disable}"
export JWT_ISSUER="${JWT_ISSUER:-narubox-bot-dev}"
export CALLBACK_ORIGIN="${CALLBACK_ORIGIN:-http://localhost:3000}"
export URI_MASTER="${URI_MASTER:-http://localhost:3000}"
export APIKEY="${APIKEY:-dummy-youtube-api-key}"

if ! command -v go >/dev/null 2>&1; then
  echo "go command not found" >&2
  exit 1
fi

if [[ -z "${JWT_SECRET:-}" ]]; then
  if ! command -v openssl >/dev/null 2>&1; then
    echo "openssl command not found" >&2
    exit 1
  fi
  export JWT_SECRET
  JWT_SECRET="$(openssl rand -hex 16)"
fi

if docker compose version >/dev/null 2>&1; then
  compose=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  compose=(docker-compose)
else
  echo "docker compose command not found" >&2
  exit 1
fi

"${compose[@]}" up -d postgres

echo "Waiting for postgres..."
for _ in {1..30}; do
  if docker exec narubox-postgres pg_isready -U narubox -d narubox_dev >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

if ! docker exec narubox-postgres pg_isready -U narubox -d narubox_dev >/dev/null 2>&1; then
  echo "postgres did not become ready in time" >&2
  exit 1
fi

go run . init

echo "Starting narubox-bot on http://localhost:3000"
exec go run .
