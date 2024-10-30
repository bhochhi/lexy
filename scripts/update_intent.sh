#!/bin/bash

BOT_ID=$1
LOCALE_ID=$2
BOT_VERSION=$3
INTENT_ID=$4
INTENT_NAME=$5
SLOTS=$6

# Retrieve the current intent configuration
CURRENT_INTENT=$(aws lexv2-models describe-intent --bot-id "$BOT_ID" --bot-version "$BOT_VERSION" --locale-id "$LOCALE_ID" --intent-id "$INTENT_ID")

# Filter out invalid fields
FILTERED_INTENT=$(echo "$CURRENT_INTENT" | jq 'del(.creationDateTime, .lastUpdatedDateTime, .version, .responseMetadata)')

# Update the slot priorities in the current intent configuration
UPDATED_INTENT=$(echo "$FILTERED_INTENT" | jq --argjson slots "$SLOTS" '.slotPriorities = $slots')

echo "Here is the updated intent: ===> $UPDATED_INTENT"
# Update the intent with the modified configuration
# Print the command being executed
echo "Executing command: aws lexv2-models update-intent --bot-id \"$BOT_ID\" --bot-version \"$BOT_VERSION\" --locale-id \"$LOCALE_ID\" --intent-id \"$INTENT_ID\" --cli-input-json \"$UPDATED_INTENT\""

RESPONSE=$(aws lexv2-models update-intent \
  --bot-id "$BOT_ID" \
  --bot-version "$BOT_VERSION" \
  --locale-id "$LOCALE_ID" \
  --intent-id "$INTENT_ID" \
  --cli-input-json "$UPDATED_INTENT")

# Print the response
echo "Here is the response from update intent: ===> $RESPONSE"