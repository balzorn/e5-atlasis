#!/usr/bin/env bash
set -euo pipefail

corepack enable
corepack prepare pnpm@12.9.1 --activate

cd /workspaces/e5-atlasis

if [ -f backend/go.mod ]; then
  cd backend
  go mod tidy
  cd ..
fi

if [ -f frontend/package.json ]; then
  cd frontend
  pnpm install
  cd ..
fi
