#!/usr/bin/env bash
set -euo pipefail

echo "[entrypoint] starting"

if command -v node >/dev/null 2>&1; then
  echo "[entrypoint] node: $(node -v)"
else
  echo "[entrypoint] node: not found"
fi

if command -v npm >/dev/null 2>&1; then
  echo "[entrypoint] npm: $(npm -v)"
else
  echo "[entrypoint] npm: not found"
fi

if command -v caddy >/dev/null 2>&1; then
  # caddy prints version with either 'caddy version' or 'caddy --version' depending on build
  if caddy version >/dev/null 2>&1; then
    echo "[entrypoint] caddy: $(caddy version)"
  else
    echo "[entrypoint] caddy: $(caddy --version 2>/dev/null || echo 'unknown')"
  fi
else
  echo "[entrypoint] caddy: not found"
fi

echo "[entrypoint] args: $*"

if [ "$#" -eq 0 ]; then
  echo "[entrypoint] no command provided — launching a shell"
  exec /bin/bash
else
  echo "[entrypoint] exec: $*"
  exec "$@"
fi
