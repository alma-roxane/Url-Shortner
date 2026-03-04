#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$ROOT_DIR/.logs"
mkdir -p "$LOG_DIR" "$ROOT_DIR/db-service/data"

pids=()
cleanup() {
  for pid in "${pids[@]:-}"; do
    kill "$pid" >/dev/null 2>&1 || true
  done
}
trap cleanup EXIT INT TERM

start_service() {
  local name="$1"
  local dir="$2"
  shift 2

  (
    cd "$ROOT_DIR/$dir"
    "$@"
  ) >"$LOG_DIR/$name.log" 2>&1 &

  local pid=$!
  pids+=("$pid")
  printf '%s started (pid=%s)\n' "$name" "$pid"
}

start_service db-service db-service env PORT=8083 DB_SNAPSHOT_PATH=./data/urls.json go run ./main.go
start_service url-service url-service env PORT=8081 DB_SERVICE_URL=http://localhost:8083 PUBLIC_BASE_URL=http://localhost:8080 go run ./main.go
start_service redirect-service redirect-service env PORT=8082 DB_SERVICE_URL=http://localhost:8083 go run ./main.go
start_service api-gateway api-gateway env PORT=8080 URL_SERVICE_URL=http://localhost:8081 REDIRECT_SERVICE_URL=http://localhost:8082 API_KEY=local-dev-key RATE_LIMIT_PER_MIN=120 go run ./main.go

echo
for i in {1..30}; do
  if curl -fsS http://localhost:8080/health >/dev/null 2>&1; then
    echo "All services are up"
    echo "Open: http://localhost:8080/app"
    echo "API key in UI default is: local-dev-key"
    wait
    exit 0
  fi
  sleep 1
done

echo "Gateway did not become healthy. Check logs in $LOG_DIR"
exit 1
