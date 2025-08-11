# Architecture & Implementation Plan (high level)

**Goal:** Modern GenAI-enabled chatbot for USAA (banking + insurance). Use Bedrock to *classify intent*, then route to Amazon Lex for full dialog handling and fulfillment. Backend in **Node.js**. All components on AWS.

* **API Gateway / Ingress**: Public API endpoint for the frontend to call.
* **Nodejs Backend** (Lambda): receives user messages, calls **Amazon Bedrock** for intent classification, Lambda routes to **Amazon Lex** bot, handles Lex dialog events, and performs fulfillment.
* **Amazon Bedrock**: Light-weight classifier to detect top-level intent & confidence score. bedrock should continue the dialog until it has enough context to classify the intent. Bedrock should have these 10 intents supported by lex.

* **Amazon Lex (V2)**: Dialog management for intents with slot elicitation, confirmations, and fulfillment from lex itself. we will do codehook later.

1. client from postman for now sends `{userId, message, sessionId}` to API Gateway.
2. Gateway invokes backend (Lambda, lets make is in nodejs).
3. Lambda  calls **Bedrock** with message to *classify intent* (returns intent + confidence).
4. If `confidence >= threshold` (configurable), backend routes message + context to corresponding **Lex bot alias** with intent hint.
   * If low confidence, bedrock should return clarifying question to user or route to human so that bedrock can help find the intent.
5. **Lex** runs dialog flow (elicits slots). Lex may call backend fulfillment webhook when intent complete.
6. postman displays response.

---

# AWS Services & Key Configuration Notes

* **Amazon Bedrock**: Use for intent classification only (fast, low-cost calls). Ensure model choice is optimized for classification (choose small instruction-tuned model if available).

* **Amazon Lex (V2)**: Create bot with 10 intents (below). Use Lex's sessionAttributes to keep context. Configure Lambda webhook for fulfillment. let's use go aws sdk for lex to creat bot, create intent and build the bot.

* **Nodejs Backend**: Provide Bedrock client calls (AWS SDK), Lex Runtime API calls (StartConversation / RecognizeText for V2 or runtime/v2), and perform pre-signed operations where needed.

* **IAM**:

  * Bedrock/ Lex/ access to backend role.
  * Lex needs permission to invoke backend fulfillment Lambda (if using Lambda).

---

# Create Ten Lex Intents (with sample utterances + suggested slots)

Below are suggested Lex intent names, short description, sample utterances, and slots (type hints):

1. **routing\_number**

   * Purpose: Provide bank routing number info.
   * Samples: “What is the routing number for my account?”, “Give me the routing number for checking.”
   * Slots: `accountType` (enum: checking, savings, other), `state` (optional)

2. **dispute\_transaction**

   * Purpose: Start a transaction dispute flow.
   * Samples: “I want to dispute a transaction”, “There’s a charge I don’t recognize from 07/15”
   * Slots: `transactionDate` (date), `amount` (number), `merchant` (string), `last4` (string), `reason` (enum/slot)

3. **credit\_rate\_interest**

   * Purpose: Provide credit card interest rate / APR info.
   * Samples: “What’s my credit card interest rate?”, “How much interest am I paying?”
   * Slots: `cardType` (enum: credit, loan), `accountId` (optional masked id)

4. **pay\_bill**

   * Purpose: Initiate bill payment.
   * Samples: “Pay my electric bill”, “Transfer \$200 to my mortgage”
   * Slots: `payee` (string), `amount` (number), `fromAccount` (enum/list), `date` (date)

5. **transfer\_funds**

   * Purpose: Transfer funds internal.
   * Samples: “Transfer \$500 to savings”, “Move money between accounts”
   * Slots: `fromAccount`, `toAccount`, `amount`, `date`

6. **report\_card\_lost**

   * Purpose: Report lost/stolen card and request block.
   * Samples: “I lost my debit card”, “Block my credit card immediately”
   * Slots: `cardType`, `last4`, `whenLost` (date/time)

7. **claim\_insurance**

   * Purpose: Start an insurance claim (auto/home).
   * Samples: “I need to file an auto claim”, “How do I file a property damage claim?”
   * Slots: `policyNumber`, `claimType` (auto/home), `incidentDate`, `description`, `location`

8. **coverage\_info**

   * Purpose: Ask about insurance coverage details.
   * Samples: “What’s covered under my policy?”, “Do I have rental reimbursement?”
   * Slots: `policyNumber`, `coverageItem` (string/enum)

9. **branch\_hours\_location**

   * Purpose: Branch/ATM hours or location info.
   * Samples: “Where is the nearest branch?”, “What time does the San Antonio branch open?”
   * Slots: `location` (city, zip), `branchName` (optional)

10. **agent\_handoff / escalate\_to\_human**

* Purpose: Route user to human agent or schedule callback.
* Samples: “I want to talk to a human”, “Schedule a callback tomorrow”
* Slots: `preferredTime`, `reason`, `contactNumber`

**Notes:** For each slot use built-in slot types (date, time, number, phoneNumber) where possible. Configure slot validation & prompts. Use slot prompts and confirmation prompts in Lex.

---

# Example Dialog Flow (dispute\_transaction)

1. User: “I want to dispute a charge.”
2. postman → Backend → Bedrock returns `dispute_transaction` (high confidence).
3. Backend calls Lex with intent hint `dispute_transaction`. Lex: “I’m sorry to hear that. What was the transaction date?”
4. User: “July 15, 2025.”
5. Lex elicits `amount`, `merchant`, and `last4`. After all slots filled, Lex asks: “Do you want me to submit this dispute now?”
6. User: “Yes.”
7. Lex calls backend fulfillment webhook with slot values. But not needed for mVP. let's just return static fulfillment response from lex.
8. Lex returns final message to user: “Your dispute has been submitted (ID ABC123). We’ll follow up.”

---

1. **Project scaffolding**

   * Create repo, README, folder layout, `.env.example`.

2. **Backend skeleton (Node.js)**

   * Bedrock client wrapper (classification).
   * Lex runtime integration (start convo, send text).
   * Simple fulfillment endpoints (mock).


6. **End-to-end**

   * User ->API gateway-->lambda--> Bedrock
   - after intent detected.
                          --->Lambda--> ---> Lex  -> response.


10. **Deployment**

* for mvp. let's focus on thesee
  * Use AWS CDK for infrastructure.
  * Deploy Lambda, API Gateway, Lex bot.
  * Configure IAM roles/policies.
  * Provide deployment script (e.g., `scripts/deploy.sh`) as needed.

---

## Implementation Notes (Added by Build)

Provisioned via CDK (LexyStack):

* Lex V2 Bot (10 intents) + Version + prod Alias
* Lambda (Node.js) for /chat classification + routing
* API Gateway REST (POST /chat)
* IAM policies (Bedrock InvokeModel, Lex runtime)
* Deployment script `scripts/deploy.sh`

Lambda env: LEX_BOT_ID, LEX_BOT_ALIAS_ID, LEX_LOCALE_ID, INTENT_CONFIDENCE_THRESHOLD, BEDROCK_MODEL_ID.

Next: add fulfillment hooks, persistence, logging improvements.

