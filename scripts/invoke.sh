#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 2 ]; then
  echo "Usage: $0 <api-url> <message> [session-id]" >&2
  exit 1
fi

API_URL="$1"
MSG="$2"
SESSION_ID="${3:-testsession}"
USER_ID="user-${USER:-local}"

curl -s -X POST "$API_URL" \
  -H 'Content-Type: application/json' \
  -d "{\"userId\":\"$USER_ID\",\"sessionId\":\"$SESSION_ID\",\"message\":\"$MSG\"}" | jq . || echo "(Install jq for pretty output)"
