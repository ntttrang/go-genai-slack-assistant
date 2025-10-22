#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if SLACK_SIGNING_SECRET is set
if [ -z "$SLACK_SIGNING_SECRET" ]; then
    echo "Error: SLACK_SIGNING_SECRET is not set. Please set it in .env file or export it."
    exit 1
fi

TIMESTAMP=$(date +%s)
BODY='{"type":"event_callback","event":{"type":"message","text":"Hello world","channel":"chatchit","user":"U123456","ts":"'${TIMESTAMP}'.000001"}}'
SIG_BASE="v0:${TIMESTAMP}:${BODY}"

# Generate HMAC signature using the same method as Slack
# Extract the hash value from openssl output (last field)
HASH=$(echo -n "$SIG_BASE" | openssl dgst -sha256 -hmac "$SLACK_SIGNING_SECRET" | awk '{print $NF}')
COMPUTED_SIG="v0=${HASH}"

echo "Testing Slack webhook endpoint..."
echo "Timestamp: $TIMESTAMP"
echo "Signature: $COMPUTED_SIG"
echo ""

curl -X POST http://localhost:8080/slack/events \
    -H "Content-Type: application/json" \
    -H "X-Slack-Request-Timestamp: $TIMESTAMP" \
    -H "X-Slack-Signature: $COMPUTED_SIG" \
    -d "$BODY"

echo ""