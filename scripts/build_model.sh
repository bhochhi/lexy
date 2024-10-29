#!/bin/bash

BOT_ID=$1
LOCALE_ID=$2
BOT_VERSION=$3

while true; do
  response=$(aws lexv2-models describe-bot-locale --bot-id "$BOT_ID" --bot-version "$BOT_VERSION" --locale-id "$LOCALE_ID")
  status=$(echo "$response" | jq -r '.botLocaleStatus')
  if [ "$status" == "Failed" ]; then
    echo "Build failed with status: \n$response"
    exit 1
  elif [ "$status" != "Building" ]; then
    echo "Build succeeded with status: \n$response"
    exit 0
  fi
  sleep 10
done