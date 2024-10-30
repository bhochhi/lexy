#!/bin/bash

BOT_ID=$1
LOCALE_ID=$2
BOT_VERSION=$3

COUNT=0
MAX_RETRIES=40  # Increased the maximum retries to allow more time for the build process


while true; do
  response=$(aws lexv2-models describe-bot-locale --bot-id "$BOT_ID" --bot-version "$BOT_VERSION" --locale-id "$LOCALE_ID")
  status=$(echo "$response" | jq -r '.botLocaleStatus')
  if [ "$status" == "Failed" ]; then
    echo "Build failed with status: \n$response"
    exit 1
  elif [ "$status" == "Ready" ]; then
    echo "Build succeeded with status: \n$response"
    exit 0
  elif [ "$COUNT" -gt $MAX_RETRIES ]; then
    echo "Build timed out with status: \n$response"
    exit 1
  else
    echo "Build in progress with status: $status"
    COUNT=$((COUNT+1))
  fi
  sleep 10
done