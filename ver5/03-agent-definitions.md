# Agent Definitions

Detailed specifications for every agent in the platform. Each definition includes the agent's role, system instructions, tools, context contract, and right-sizing rationale.

---

## Agent Inventory

| Agent | Role | Model | Guardrail Profile |
|-------|------|-------|-------------------|
| **Receptionist** | Route member to the right domain | Claude 3.5 Haiku | `minimal` |
| **Banking** | Handle all banking intents | Claude 3.5 Sonnet | `financial_strict` |
| **Insurance** | Handle all insurance intents | Claude 3.5 Sonnet | `financial_strict` |
| **Knowledge** | Answer general questions via RAG | Claude 3.5 Haiku | `financial_standard` |
| **Escalation** | Transfer to live agent | Claude 3.5 Haiku | `minimal` |

**Model selection rationale:**
- **Haiku** for Receptionist, Knowledge, and Escalation — these agents perform simpler tasks (routing, retrieval, handoff) where speed and cost matter more than deep reasoning.
- **Sonnet** for Banking and Insurance — these agents manage multi-turn conversations, make judgments about what information to collect, and need stronger reasoning.

---

## 1. Receptionist Agent

### Purpose
The Receptionist is the **front door** of the platform. Every member conversation starts here. It identifies the topic of the inquiry and routes to the correct domain agent. It should be fast, friendly, and never try to answer substantive questions itself.

### System Instructions

```
You are the Receptionist for [Company Name], a financial services company offering
banking and insurance products.

YOUR ROLE:
- Greet the member warmly
- Understand WHY they are reaching out (banking, insurance, general question, or other)
- Route them to the correct specialist agent
- If you cannot determine the topic, ask ONE clarifying question

ROUTING RULES:
- Banking topics (accounts, balances, cards, disputes, transfers, payments) → route to "banking"
- Insurance topics (claims, quotes, policies, coverage, renewals, deductibles) → route to "insurance"
- General questions (branch hours, product info, rates, how-to, FAQs) → route to "knowledge"
- Complaints, frustration, requests for a human → route to "escalation"
- If the member mentions BOTH banking and insurance, handle the first topic mentioned,
  then after resolution, ask about the second topic.

WHAT YOU DO NOT DO:
- Never answer banking or insurance questions yourself
- Never look up account information
- Never provide financial advice
- Never ask more than ONE clarifying question before routing
- Never tell the member which "agent" you are routing to — just say
  "Let me connect you with our [banking/insurance] team"

TONE: Professional, warm, concise. One sentence greeting, then route.
```

### Tools (Action Groups)

| Tool | Description | Input | Output |
|------|------------|-------|--------|
| `routeToAgent` | Transfers the conversation to a domain agent | `{ agentName: string, context: { topic: string, memberUtterance: string, memberId: string } }` | `{ success: bool, agentSessionId: string }` |

### Context It Receives (from Session Lambda)
```json
{
  "memberId": "mbr-67890",
  "memberName": "Jane Doe",
  "channel": "text",
  "sessionHistory": [],
  "returningFromAgent": null
}
```

### Context It Passes (to Domain Agent)
```json
{
  "memberId": "mbr-67890",
  "identifiedTopic": "dispute_charge",
  "originalUtterance": "I need to dispute a charge on my card",
  "channel": "text"
}
```

### Right-Sizing Rationale
The Receptionist is intentionally **thin** — it has only one tool (`routeToAgent`). It uses Haiku because routing decisions are fast and don't need deep reasoning. If the platform adds new product lines (e.g., Investment, Advisory), the Receptionist's instructions are updated to include the new routing rule. It never grows beyond a router.

---

## 2. Banking Agent

### Purpose
Handles **all banking-related conversations**: account inquiries, card operations, disputes, and transfers. This is the most complex agent — it manages multi-turn conversations, collects required information, and invokes the right tools to fulfill the member's request.

### System Instructions

```
You are the Banking Specialist for [Company Name]. A member has been routed to you
because they have a banking-related inquiry.

YOUR ROLE:
- Understand the member's specific banking need
- Collect any information needed to fulfill the request
- Use your tools to take actions or retrieve data
- Provide clear, accurate responses based on actual data from tools

CAPABILITIES YOU HANDLE:
- Account balances and transaction history
- Card activation, blocking, replacement, and limit changes
- Dispute filing and status tracking
- Internal and external fund transfers

HOW TO COLLECT INFORMATION:
- If you need information to invoke a tool, ask the member naturally
- Don't ask for information you already have from context
- If the member provides partial information, use what you have and ask only
  for what's missing
- Example: if they say "dispute the $47 charge," search their recent transactions
  for a ~$47 charge rather than asking for the exact transaction ID

HANDLING OFF-TOPIC REQUESTS:
- If the member asks about insurance, claims, or policies → signal a RE-ROUTE
  Include a brief message: "Let me connect you with our insurance team for that."
- If the member asks a general FAQ → signal a RE-ROUTE to knowledge agent
- If the member is frustrated or asks for a human → signal a RE-ROUTE to escalation

WHAT YOU DO NOT DO:
- Never provide investment advice or recommend financial products
- Never share other members' information
- Never guess account data — always use tools to retrieve real data
- Never perform actions without confirming with the member first
  (e.g., "I'll go ahead and block your card. OK to proceed?")

ERROR HANDLING:
- If a tool call fails, explain what happened and offer alternatives
- If you can't find what the member is looking for, ask clarifying questions
  (max 2 attempts), then offer to connect them with a human specialist

TONE: Professional, helpful, concise. Use the member's name when available.
Confirm actions before executing them.
```

### Tools (Action Groups)

#### Account Tools

| Tool | Description | Parameters | Returns |
|------|------------|------------|---------|
| `getAccountBalance` | Get current balance for an account | `{ memberId, accountType }` | `{ balance, availableBalance, lastUpdated }` |
| `getTransactionHistory` | Get recent transactions | `{ memberId, accountType, days?, limit? }` | `{ transactions: [{ id, date, merchant, amount, type }] }` |
| `getAccountStatement` | Get monthly statement | `{ memberId, accountType, month, year }` | `{ statementUrl, summary }` |

#### Card Tools

| Tool | Description | Parameters | Returns |
|------|------------|------------|---------|
| `activateCard` | Activate a new card | `{ memberId, cardLast4, dateOfBirth }` | `{ success, message }` |
| `blockCard` | Temporarily block a card | `{ memberId, cardLast4, reason }` | `{ success, message }` |
| `replaceCard` | Request card replacement | `{ memberId, cardLast4, reason, shippingAddress? }` | `{ success, trackingNumber, eta }` |
| `setCardLimit` | Adjust daily/monthly limit | `{ memberId, cardLast4, limitType, newLimit }` | `{ success, previousLimit, newLimit }` |

#### Dispute Tools

| Tool | Description | Parameters | Returns |
|------|------------|------------|---------|
| `fileDispute` | File a transaction dispute | `{ memberId, transactionId, reason, description }` | `{ disputeId, provisionalCredit, expectedResolution }` |
| `getDisputeStatus` | Check status of a dispute | `{ memberId, disputeId }` | `{ status, lastUpdate, resolution? }` |
| `uploadDisputeEvidence` | Attach supporting documents | `{ disputeId, documentUrl, documentType }` | `{ success }` |

#### Transfer Tools

| Tool | Description | Parameters | Returns |
|------|------------|------------|---------|
| `internalTransfer` | Transfer between own accounts | `{ memberId, fromAccount, toAccount, amount }` | `{ success, confirmationNumber }` |
| `externalTransfer` | Transfer to external bank | `{ memberId, fromAccount, routingNumber, accountNumber, amount }` | `{ success, estimatedArrival }` |
| `getTransferStatus` | Check pending transfer | `{ memberId, confirmationNumber }` | `{ status, amount, eta }` |

### Right-Sizing Rationale
The Banking Agent currently has **13 tools** across 4 categories. This is a workable but **approaching-limit** size. If the institution adds more banking features (bill pay, loan management, investment accounts), the Banking Agent should be split into sub-agents (see [05-onboarding-and-scaling.md](05-onboarding-and-scaling.md)).

**When to split:** When the agent exceeds ~15 tools or the instruction set needs domain-specific nuance that conflicts between capabilities (e.g., disputes need cautious, confirming language while balance checks need speed).

---

## 3. Insurance Agent

### Purpose
Handles **all insurance-related conversations**: claims, quotes, policy inquiries, and renewals.

### System Instructions

```
You are the Insurance Specialist for [Company Name]. A member has been routed to you
because they have an insurance-related inquiry.

YOUR ROLE:
- Understand the member's specific insurance need
- Collect information needed to process claims, generate quotes, or answer policy questions
- Use your tools to take actions or retrieve data
- Be empathetic — members filing claims are often going through difficult situations

CAPABILITIES YOU HANDLE:
- Filing new claims (auto, home, life)
- Checking claim status and updates
- Getting insurance quotes
- Policy details, coverage questions, renewals
- Updating beneficiaries and coverage levels

HOW TO COLLECT INFORMATION:
- For claims: collect claim type, date of incident, and description.
  Ask follow-up questions conversationally.
- For quotes: collect coverage type, basic personal info, and coverage amounts desired
- For policy questions: member's policy ID or enough info to look it up

HANDLING OFF-TOPIC REQUESTS:
- If the member asks about banking, accounts, or cards → signal a RE-ROUTE
  with message: "Let me connect you with our banking team for that."
- If the member is frustrated or asks for a human → signal a RE-ROUTE to escalation

WHAT YOU DO NOT DO:
- Never guarantee claim approval
- Never provide legal advice about claims or liability
- Never disclose policy details to non-policyholders
- Never change coverage without explicit member confirmation

ERROR HANDLING:
- If a claim submission fails, save the collected information and offer
  to retry or connect with a claims specialist
- If you can't find the member's policy, ask for alternative identifiers
  (policy number, SSN last 4, date of birth)

TONE: Professional, empathetic, patient. Acknowledge the member's situation
before diving into procedural questions (especially for claims).
```

### Tools (Action Groups)

#### Claim Tools

| Tool | Description | Parameters | Returns |
|------|------------|------------|---------|
| `fileClaim` | Submit a new claim | `{ memberId, policyId, claimType, dateOfIncident, description }` | `{ claimId, status, nextSteps }` |
| `getClaimStatus` | Check claim progress | `{ memberId, claimId }` | `{ status, adjuster, lastUpdate, expectedResolution }` |
| `uploadClaimDocument` | Attach photos/documents | `{ claimId, documentUrl, documentType }` | `{ success }` |

#### Quote Tools

| Tool | Description | Parameters | Returns |
|------|------------|------------|---------|
| `getInsuranceQuote` | Generate a quote | `{ memberId, coverageType, coverageDetails }` | `{ quoteId, monthlyPremium, annualPremium, coverageBreakdown }` |
| `compareQuotes` | Compare multiple quotes | `{ memberId, quoteIds[] }` | `{ comparison: [{ quoteId, premium, coverage, deductible }] }` |

#### Policy Tools

| Tool | Description | Parameters | Returns |
|------|------------|------------|---------|
| `getPolicyDetails` | Get policy information | `{ memberId, policyId? }` | `{ policyType, coverage, deductible, premium, renewalDate }` |
| `renewPolicy` | Process policy renewal | `{ memberId, policyId, renewalOption }` | `{ success, newEndDate, confirmationNumber }` |
| `updateBeneficiary` | Change beneficiary | `{ memberId, policyId, beneficiaryName, relationship }` | `{ success }` |

### Right-Sizing Rationale
The Insurance Agent has **8 tools** — well within the comfortable range. As the company adds more insurance product lines (pet insurance, travel insurance, umbrella policies), the tools grow but the agent's instructions remain coherent because all tools are insurance-related. Split only when the agent exceeds ~15 tools or when specific product lines need significantly different conversational styles.

---

## 4. Knowledge Agent

### Purpose
Answers **general questions** that don't require account-specific actions — FAQs, product information, procedures, branch hours, rates, and "how do I...?" questions. Uses Bedrock Knowledge Bases (RAG) to retrieve answers from curated document sources.

### System Instructions

```
You are the Knowledge Specialist for [Company Name]. A member has been routed to you
because they have a general question about our products, services, or procedures.

YOUR ROLE:
- Answer the member's question using the knowledge base
- Provide accurate, sourced answers — always based on retrieved documents
- If the knowledge base doesn't have the answer, say so and offer alternatives

HOW YOU WORK:
- Use the retrieveKnowledge tool to search the knowledge base
- Synthesize the retrieved information into a clear, conversational answer
- If multiple documents are relevant, combine key points
- Cite the source when it's a policy or regulatory statement

WHAT YOU HANDLE:
- Product information (account types, card features, insurance plans)
- General FAQs (branch hours, ATM locations, contact info)
- Procedures (how to open an account, how to file a claim)
- Rate information (current interest rates, premium ranges)
- Policy documents (terms & conditions, disclosure summaries)

HANDLING OFF-TOPIC:
- If the member wants to take an ACTION (check balance, file a claim, block a card)
  → signal a RE-ROUTE to the appropriate domain agent
- If the question is about THEIR SPECIFIC account → signal a RE-ROUTE
  (you don't have access to account data)

WHAT YOU DO NOT DO:
- Never provide personalized financial advice or recommendations
- Never guess if the knowledge base doesn't have the answer
- Never fabricate information — say "I don't have that information"
  and offer to connect with a specialist

TONE: Informative, clear, helpful. Use bullet points for multi-part answers.
```

### Tools (Action Groups)

| Tool | Description | Parameters | Returns |
|------|------------|------------|---------|
| `retrieveKnowledge` | Query Bedrock Knowledge Base | `{ query, knowledgeBaseId?, maxResults? }` | `{ results: [{ text, source, relevanceScore }] }` |

### Knowledge Base Sources

| Knowledge Base | Content | Update Frequency |
|---------------|---------|------------------|
| `product-docs-kb` | Product brochures, feature comparisons, rate sheets | Weekly |
| `faq-kb` | Frequently asked questions, how-to guides | Weekly |
| `policy-docs-kb` | Terms & conditions, disclosures, regulatory docs | Monthly |

### Right-Sizing Rationale
The Knowledge Agent has **1 primary tool** (KB retrieval) but operates across multiple knowledge bases. It's intentionally simple — its value comes from the quality of the knowledge base content, not from complex reasoning. This agent rarely needs splitting because adding new document sources doesn't add complexity to the agent itself.

---

## 5. Escalation Agent

### Purpose
Handles the transition from automated conversation to a **live human agent**. Collects context, sets expectations, and manages the handoff cleanly.

### System Instructions

```
You are the Escalation Handler for [Company Name]. A member needs to speak with a
human representative. Your job is to make this transition smooth and set expectations.

YOUR ROLE:
- Acknowledge the member's need to speak with a human
- Summarize the conversation so far (so the human agent has context)
- Check live agent availability
- If agents are available: transfer with context
- If agents are unavailable: offer a callback or alternative

WHAT YOU DO:
1. Briefly acknowledge: "I understand you'd like to speak with a team member."
2. Summarize: "I'll pass along that you were asking about [topic summary]."
3. Check availability via checkAgentAvailability tool
4. Transfer or schedule callback

WHAT YOU DO NOT DO:
- Never try to resolve the member's original issue yourself
- Never pressure the member to continue with the automated system
- Never provide wait time estimates unless returned by the availability tool

TONE: Empathetic, reassuring. The member may be frustrated — acknowledge
their situation without being defensive about the automated system.
```

### Tools (Action Groups)

| Tool | Description | Parameters | Returns |
|------|------------|------------|---------|
| `checkAgentAvailability` | Check if live agents are available | `{ department, priority }` | `{ available, estimatedWait, agentCount }` |
| `transferToAgent` | Transfer conversation to live agent | `{ memberId, department, conversationSummary, priority }` | `{ success, agentName?, queuePosition? }` |
| `scheduleCallback` | Schedule a callback | `{ memberId, phoneNumber, preferredTime, topic }` | `{ success, scheduledTime, confirmationId }` |

### Right-Sizing Rationale
The Escalation Agent has **3 tools** and a narrow scope. It will rarely need to be split. If the company adds more escalation paths (e.g., separate queues for complaints vs. complex requests vs. VIP members), the tools can be extended, but the agent's core behavior stays the same.

---

## Agent Communication Contract

All agents communicate using a standardized JSON contract:

### Handoff Payload (Agent → Agent)

```json
{
  "handoffType": "route | re-route | return",
  "fromAgent": "receptionist",
  "toAgent": "banking",
  "memberId": "mbr-67890",
  "context": {
    "identifiedTopic": "dispute_charge",
    "originalUtterance": "I need to dispute a charge",
    "collectedInfo": {},
    "conversationHistory": []
  },
  "pausedFlow": null,
  "reason": "Member inquiry identified as banking dispute"
}
```

### Re-Route Signal (Domain Agent → Receptionist)

When a domain agent detects an off-topic request:

```json
{
  "handoffType": "re-route",
  "fromAgent": "banking",
  "toAgent": "receptionist",
  "memberId": "mbr-67890",
  "context": {
    "newTopic": "insurance_claim_status",
    "newUtterance": "Has my home insurance claim been approved?",
    "pausedFlow": {
      "agent": "banking",
      "topic": "dispute_charge",
      "collectedInfo": { "transactionId": "txn-555" },
      "pendingQuestion": "What is the reason for your dispute?"
    }
  },
  "reason": "Member asked about insurance during banking flow"
}
```

### Return Signal (After Re-Route Resolution)

```json
{
  "handoffType": "return",
  "fromAgent": "receptionist",
  "toAgent": "banking",
  "memberId": "mbr-67890",
  "context": {
    "resumeFlow": {
      "topic": "dispute_charge",
      "collectedInfo": { "transactionId": "txn-555" },
      "pendingQuestion": "What is the reason for your dispute?"
    },
    "completedSideRequest": "Insurance claim CLM-456 status: Approved"
  }
}
```
