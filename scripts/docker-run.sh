#!/usr/bin/env bash
set -euo pipefail

IMAGE_NAME="param-formatter:local"
CONTAINER_NAME="param-formatter"
PORT="8080"

echo "[1/3] Building image: ${IMAGE_NAME}"
docker build -t "${IMAGE_NAME}" .

echo "[2/3] Removing old container if exists: ${CONTAINER_NAME}"
docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true

echo "[3/3] Starting container: ${CONTAINER_NAME}"
docker run -d --name "${CONTAINER_NAME}" -p "${PORT}:8080" "${IMAGE_NAME}"

echo "Container started. Try:"
echo "  curl -X POST http://127.0.0.1:${PORT}/param_formatter \\
    -H 'Content-Type: application/json' \\
    -d '{"backup_type":"db"}'"
echo "  curl http://127.0.0.1:${PORT}/schema"
