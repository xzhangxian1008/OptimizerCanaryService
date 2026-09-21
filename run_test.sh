#!/usr/bin/env bash

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly BINARY="${SCRIPT_DIR}/diagnostic-service"
readonly CONFIG_FILE="${SCRIPT_DIR}/config.example.toml"
readonly HTTP_ADDR='127.0.0.1:8080'
readonly BASE_URL="http://${HTTP_ADDR}"

if [[ ! -x "${BINARY}" ]]; then
  echo "diagnostic-service is missing; run 'go build -o diagnostic-service ./cmd' first" >&2
  exit 1
fi

LOG_FILE="$(mktemp)"

cleanup() {
  if [[ -n "${SERVICE_PID:-}" ]]; then
    kill "${SERVICE_PID}" 2>/dev/null || true
    wait "${SERVICE_PID}" 2>/dev/null || true
  fi
  rm -f "${LOG_FILE}"
}
trap cleanup EXIT INT TERM

"${BINARY}" \
  -config "${CONFIG_FILE}" \
  -http-addr "${HTTP_ADDR}" >"${LOG_FILE}" 2>&1 &
SERVICE_PID=$!

for _ in {1..50}; do
  if curl --silent --fail "${BASE_URL}/healthz" >/dev/null; then
    break
  fi
  if ! kill -0 "${SERVICE_PID}" 2>/dev/null; then
    echo "diagnostic-service exited before becoming ready:" >&2
    cat "${LOG_FILE}" >&2
    exit 1
  fi
  sleep 0.1
done

if ! curl --silent --fail "${BASE_URL}/healthz" >/dev/null; then
  echo "diagnostic-service did not become ready:" >&2
  cat "${LOG_FILE}" >&2
  exit 1
fi

HTTP_RESPONSE="$(curl --silent --show-error \
  --request POST \
  "${BASE_URL}/validate")"

cat "${LOG_FILE}"
printf '%s\n' "${HTTP_RESPONSE}"
