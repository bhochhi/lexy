# Lexy Chatbot (Bedrock + Lex V2)

Minimal GenAI-enabled intent-classifying chatbot backend using:

* Amazon Bedrock (intent classification)
* Amazon Lex V2 (dialog management with 10 intents)
* AWS Lambda (Node.js 18) + API Gateway (REST) for a single /chat endpoint
* AWS CDK (TypeScript) infrastructure-as-code

## High-Level Flow

1. Client (Postman) POSTs `{ userId, sessionId, message }` to the API endpoint.
2. Lambda calls Bedrock small model to classify intent & confidence.
3. If confidence >= threshold, route to Lex RecognizeText with intent hint.
4. Lex manages slot elicitation and returns messages.
5. Lambda returns JSON containing messages for the client to display.
6. If low confidence, Lambda returns a clarification prompt (no Lex call yet).

## Intents Implemented

`routing_number, dispute_transaction, credit_rate_interest, pay_bill, transfer_funds, report_card_lost, claim_insurance, coverage_info, branch_hours_location, agent_handoff`.

## Deploy

Prereqs: Node.js 18+, AWS credentials (default profile), CDK bootstrapped account.

```bash
./scripts/deploy.sh
```

Outputs show `ApiUrl` you can call via Postman.

If Bedrock model access for `amazon.nova-micro-v1` (or another model you choose) is not yet granted, visit the Bedrock console -> Model access and enable it first, then redeploy/try again. If you cannot enable any model yet, set `USE_BEDROCK=false` in the Lambda environment (or adjust in `lib/lexy-stack.ts`) to use the built-in keyword fallback classifier.

## Quick Test (CLI)

After deploy, capture the `ApiUrl` output (looks like `https://xxxxxxxx.execute-api.<region>.amazonaws.com/prod/chat`). Then:

```bash
./scripts/invoke.sh https://xxxxxxxx.execute-api.<region>.amazonaws.com/prod/chat "I lost my debit card"
```

Sample output:

```json
{
  "type": "lex",
  "intent": "report_card_lost",
  "confidence": 0.82,
  "messages": ["I'm sorry to hear that. When did you lose it?"]
}
```

Try a second message continuing the same session (supply same session id):

```bash
./scripts/invoke.sh https://xxxxxxxx.execute-api.<region>.amazonaws.com/prod/chat "Yesterday afternoon" testsession
```

If classification is low confidence you'll get `type=clarification`.

## Example Request (Postman)

POST ApiUrl

Body (raw JSON):

```json
{
  "userId": "user123",
  "sessionId": "session123",
  "message": "I lost my debit card"
}
```

## Example Response

```json
{
  "type": "lex",
  "intent": "report_card_lost",
  "confidence": 0.93,
  "messages": ["I'm sorry to hear that. When did you lose it?"]
}
```

If low confidence:

```json
{ "type": "clarification", "message": "Could you clarify what you need help with?" }
```

## Local Build

```bash
npm install
npm run build
```

## Cleanup

Basic destroy:

```bash
npm run destroy
```

Or use the helper (supports deleting bootstrap stack, confirmations, custom region):

```bash
./scripts/undeploy.sh --region us-east-1
```

Full cleanup including bootstrap (only if this environment is disposable):

```bash
./scripts/undeploy.sh --remove-bootstrap --force --region us-east-1
```

If destroy fails due to Lex bot dependencies, re-run after a minute (Lex version deletion can be eventual). You can also manually delete the CloudFormation stack in the console.

## Notes / Next Steps

* Add fulfillment Lambda hooks per intent.
* Persist conversation state / transcripts (e.g., DynamoDB).
* Add structured logging & tracing (X-Ray).
* Introduce guardrails / moderation before Bedrock call.
* Parameterize model via CDK context (e.g. `cdk deploy -c modelId=amazon.nova-micro-v1`).

### Changing / Disabling Bedrock Model

Edit `BEDROCK_MODEL_ID` in `lib/lexy-stack.ts` (or `.env.example` for reference). To disable Bedrock temporarily (use keyword fallback intent classifier), set `USE_BEDROCK=false` and redeploy:

```bash
sed -i '' "s/USE_BEDROCK: 'true'/USE_BEDROCK: 'false'/" lib/lexy-stack.ts
npm run build
./scripts/deploy.sh
```
