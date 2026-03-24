# Orchestration & Recovery Patterns

How agents hand off conversations, handle unexpected member inputs, recover from errors, and manage mid-flow topic switches. This document covers both the **happy path** and every recovery scenario.

---

## The Orchestration Loop

Every member message flows through this loop:

```mermaid
flowchart TD
    Msg["Member sends message"] --> SL["Session Lambda\n(load session state)"]
    SL --> Active{"Active agent\nin session?"}

    Active -->|"No (new conversation)"| RA["Receptionist Agent"]
    Active -->|"Yes"| DA["Route to active\nDomain Agent"]

    RA --> Route{"Topic\nidentified?"}
    Route -->|"Yes"| DA
    Route -->|"Unclear"| Clarify["Receptionist asks\nONE clarifying question"]
    Clarify --> RA

    DA --> Process{"Agent processes\nmessage"}

    Process -->|"Normal flow"| Tool["Invoke tool\n(if needed)"]
    Process -->|"Needs more info"| Ask["Ask member\nfor information"]
    Process -->|"Off-topic detected"| Reroute["Signal RE-ROUTE\nto Receptionist"]
    Process -->|"Can't understand"| Recover["Recovery\nPattern"]
    Process -->|"Member wants human"| Esc["Route to\nEscalation Agent"]

    Tool --> Respond["Build response\n+ save session"]
    Ask --> Respond
    Reroute --> RA
    Recover --> DA
    Esc --> EA["Escalation Agent"]

    Respond --> SL2["Session Lambda\n(save state, return response)"]

    style RA fill:#2c3e50,color:#fff
    style DA fill:#2980b9,color:#fff
    style Recover fill:#e67e22,color:#fff
    style Esc fill:#e74c3c,color:#fff
```

---

## 1. Happy Path — Zero-Shot (FAQ)

The simplest case. Member asks a question, Knowledge Agent answers in one turn.

```mermaid
sequenceDiagram
    participant M as 👤 Member
    participant SL as Session Lambda
    participant RA as Receptionist Agent
    participant KA as Knowledge Agent
    participant KB as Bedrock KB

    M->>SL: "What are your branch hours?"
    SL->>RA: Route (no active agent)
    RA->>RA: Topic = general question
    RA->>KA: Route with context

    KA->>KB: retrieveKnowledge("branch hours")
    KB-->>KA: "Mon-Fri 9am-5pm, Sat 9am-1pm"
    KA-->>SL: "Our branches are open Monday through Friday, 9am to 5pm, and Saturdays 9am to 1pm."
    SL-->>M: Display response

    Note over M, SL: ✅ Resolved in one turn. No multi-turn needed.
```

**Cost:** 2 agent invocations (Receptionist Haiku + Knowledge Haiku + KB query). ~$0.001.

---

## 2. Happy Path — Multi-Turn (Action)

Member wants to take an action that requires information collection.

```mermaid
sequenceDiagram
    participant M as 👤 Member
    participant SL as Session Lambda
    participant RA as Receptionist Agent
    participant BA as Banking Agent
    participant Tool as Banking Tools

    M->>SL: "I need to dispute a charge"
    SL->>RA: Route (no active agent)
    RA->>RA: Topic = banking (dispute)
    RA-->>SL: "Let me connect you with our banking team."
    RA->>BA: Route with context: {topic: "dispute"}

    Note over SL: Session: activeAgent = "banking"

    BA-->>SL: "I can help with that. Can you describe the charge you'd like to dispute?"
    SL-->>M: Display

    M->>SL: "There's a $47.99 charge from yesterday I didn't make"
    SL->>BA: (active agent = banking)

    BA->>Tool: getTransactionHistory(memberId, days=3)
    Tool-->>BA: [{id: "txn-001", merchant: "Online Store", amount: 47.99, date: "2026-03-10"}]

    BA-->>SL: "I found a $47.99 charge at Online Store on March 10. Is this the one?"
    SL-->>M: Display

    M->>SL: "Yes, that's the one"
    SL->>BA: Continue

    BA-->>SL: "I'll file a dispute for unauthorized use. You'll receive a provisional credit within 48 hours. Shall I proceed?"
    SL-->>M: Display

    M->>SL: "Yes, go ahead"
    SL->>BA: Continue

    BA->>Tool: fileDispute(txn-001, reason="unauthorized")
    Tool-->>BA: {disputeId: "DSP-1234", provisionalCredit: true}

    BA-->>SL: "Done! Dispute DSP-1234 has been filed. A provisional credit of $47.99 will appear in your account within 48 hours. Is there anything else I can help with?"
    SL-->>M: Display

    Note over M, SL: ✅ Multi-turn resolved. 5 turns total.
```

---

## 3. Mid-Flow Topic Switch

The member changes topics during a conversation. This is the most important recovery pattern — it must feel seamless.

### Scenario: Banking → Insurance → Resume Banking

```mermaid
sequenceDiagram
    participant M as 👤 Member
    participant SL as Session Lambda
    participant RA as Receptionist Agent
    participant BA as Banking Agent
    participant IA as Insurance Agent

    M->>SL: "I need to dispute a charge"
    SL->>RA: Route
    RA->>BA: Route to banking
    BA-->>M: "Which charge would you like to dispute?"

    M->>BA: "The $47.99 from yesterday"
    BA->>BA: Found transaction. Collecting reason.
    BA-->>M: "Got it. What's the reason for your dispute?"

    M->>BA: "Wait — has my home insurance claim been approved yet?"

    Note over BA: ⚠️ OFF-TOPIC DETECTED<br/>This is an insurance question,<br/>not a banking dispute.

    BA->>RA: RE-ROUTE signal
    Note right of BA: {<br/>  handoffType: "re-route",<br/>  pausedFlow: {<br/>    topic: "dispute",<br/>    collected: {txnId: "txn-001"},<br/>    pending: "dispute reason"<br/>  },<br/>  newTopic: "insurance_claim_status"<br/>}

    RA->>IA: Route to insurance with context
    IA->>IA: getClaimStatus(memberId)
    IA-->>M: "Your home insurance claim CLM-456 was approved on March 8. The payout of $12,500 is being processed. You should receive it within 5-7 business days."

    Note over SL: Insurance question answered.<br/>Check for paused flow.

    RA-->>M: "Now, back to your dispute — you were disputing the $47.99 charge at Online Store. What's the reason for the dispute?"

    RA->>BA: RETURN with paused flow context

    M->>BA: "I didn't make that purchase"
    BA->>BA: fileDispute(txn-001, "unauthorized")
    BA-->>M: "Dispute DSP-1234 filed. Provisional credit within 48 hours."

    Note over M, SL: ✅ Both topics handled.<br/>Banking flow paused, insurance answered,<br/>banking resumed seamlessly.
```

### How It Works — The Re-Route Protocol

```mermaid
flowchart TD
    Detect["Domain Agent detects\noff-topic utterance"] --> Signal["Send RE-ROUTE signal\nto Receptionist"]
    Signal --> Pause["Pause current flow\n(save state to session)"]
    Pause --> Route["Receptionist routes\nto correct agent"]
    Route --> Resolve["New agent resolves\nthe side request"]
    Resolve --> Check{"Paused flow\nexists?"}
    Check -->|"Yes"| Resume["Receptionist resumes\npaused flow with context"]
    Check -->|"No"| Done["Ask: anything else?"]
    Resume --> Continue["Original agent continues\nfrom where it left off"]

    style Detect fill:#e67e22,color:#fff
    style Resume fill:#27ae60,color:#fff
```

---

## 4. Incomprehensible or Ambiguous Input

When the agent can't understand what the member is saying.

### Recovery Ladder

```mermaid
flowchart TD
    Unclear["Agent can't understand\nmember's message"] --> A1

    A1["Attempt 1: Ask for clarification\n'Could you rephrase that?'"] -->|"Member clarifies"| Success["✅ Resume flow"]
    A1 -->|"Still unclear"| A2

    A2["Attempt 2: Offer specific options\n'Are you looking to A, B, or C?'"] -->|"Member selects"| Success
    A2 -->|"Still unclear"| A3

    A3["Attempt 3: Re-route to Receptionist\n'Let me redirect you to make sure\nyou get the right help.'"] -->|"Receptionist re-classifies"| Success
    A3 -->|"Receptionist also can't classify"| A4

    A4["Attempt 4: Escalate to human\n'Let me connect you with\na team member.'"] --> Human["🤝 Live Agent\n(with full conversation context)"]

    style A1 fill:#3498db,color:#fff
    style A2 fill:#e67e22,color:#fff
    style A3 fill:#e67e22,color:#fff
    style A4 fill:#e74c3c,color:#fff
    style Success fill:#27ae60,color:#fff
```

### Example: Ambiguous Input During Information Collection

```mermaid
sequenceDiagram
    participant M as 👤 Member
    participant BA as Banking Agent

    BA-->>M: "Which account — checking or savings?"
    M->>BA: "the one I always use"

    Note over BA: 🤔 Ambiguous — can't determine account type

    BA-->>M: "I want to make sure I check the right account. Do you mean your checking account or your savings account?"

    M->>BA: "the one my paycheck goes to"

    Note over BA: 🧠 "Paycheck" strongly implies checking account<br/>LLM reasoning resolves this

    BA->>BA: Resolve: checking (high confidence)
    BA-->>M: "Your checking account balance is $4,250.00."

    Note over M, BA: ✅ Agent resolved ambiguity<br/>using LLM reasoning naturally
```

---

## 5. Tool Execution Failure

When a backend API or Lambda fails.

```mermaid
sequenceDiagram
    participant M as 👤 Member
    participant BA as Banking Agent
    participant Tool as Banking Tool
    participant API as Backend API

    M->>BA: "Block my card ending in 4567"
    BA-->>M: "I'll block your card ending in 4567 right away. One moment..."

    BA->>Tool: blockCard(memberId, "4567")
    Tool->>API: POST /cards/block
    API-->>Tool: 503 Service Unavailable

    Tool-->>BA: ERROR: Backend service temporarily unavailable

    Note over BA: ⚠️ Tool failed. Don't just say "error."<br/>Explain what happened and offer alternatives.

    BA-->>M: "I'm sorry — our card services system is temporarily unavailable. I wasn't able to block your card just now. Here's what I can do:"
    BA-->>M: "1. I can try again in a moment\n2. I can connect you with a team member who can block it immediately\n3. You can block your card in the mobile app under Card Settings\n\nWhich would you prefer?"

    M->>BA: "Try again"
    BA->>Tool: blockCard(memberId, "4567")
    Tool->>API: POST /cards/block
    API-->>Tool: 200 OK

    BA-->>M: "Your card ending in 4567 has been blocked. No new charges will go through. Would you like to order a replacement card?"

    Note over M, BA: ✅ Graceful recovery with alternatives
```

---

## 6. Member Explicitly Requests a Human

At any point, if the member asks for a human, the current agent immediately defers to the Escalation Agent.

```mermaid
sequenceDiagram
    participant M as 👤 Member
    participant BA as Banking Agent
    participant RA as Receptionist
    participant EA as Escalation Agent

    M->>BA: "I just want to talk to a real person"

    Note over BA: Member requests human.<br/>Don't resist — route immediately.

    BA->>RA: RE-ROUTE to escalation
    Note right of BA: {<br/>  reason: "member_requested_human",<br/>  conversationSummary: "Member was<br/>  disputing $47.99 charge, had<br/>  identified the transaction,<br/>  was about to provide reason",<br/>  pausedFlow: {...}<br/>}

    RA->>EA: Route to escalation with context
    EA->>EA: checkAgentAvailability("banking")

    alt Agents available
        EA-->>M: "I understand you'd like to speak with a team member. I'm connecting you now — and I'll pass along that you were working on a dispute so they can pick up where we left off."
        EA->>EA: transferToAgent(memberId, summary)
    else No agents available
        EA-->>M: "Our banking team is currently assisting other members. I can schedule a callback — when would be a good time?"
    end
```

---

## 7. Returning Member (Session Continuity)

If a member returns within the session window, the system picks up the conversation.

```mermaid
sequenceDiagram
    participant M as 👤 Member
    participant SL as Session Lambda
    participant DDB as DynamoDB
    participant BA as Banking Agent

    M->>SL: "Hi, I'm back"

    SL->>DDB: Load session for memberId
    DDB-->>SL: Session found: activeAgent=banking, pausedFlow=dispute

    Note over SL: Member has an active session<br/>with a paused banking dispute flow.

    SL->>BA: Resume with paused flow context

    BA-->>M: "Welcome back! Last time, you were disputing a $47.99 charge at Online Store. Would you like to continue with that?"

    M->>BA: "Yes"
    BA-->>M: "Great — what's the reason for the dispute?"

    Note over M, BA: ✅ Seamless session continuity
```

---

## Orchestration State Machine — Complete View

Summary of all scenarios and how the system handles them:

| # | Scenario | What Happens | Who Handles It |
|---|----------|-------------|----------------|
| 1 | **FAQ (zero-shot)** | Receptionist → Knowledge Agent → Answer | Knowledge Agent |
| 2 | **Multi-turn action** | Receptionist → Domain Agent → Collect info → Tool → Response | Domain Agent |
| 3 | **Topic switch mid-flow** | Domain Agent detects off-topic → RE-ROUTE → Resolve → Resume | Receptionist (coordinator) |
| 4 | **Ambiguous input** | Agent uses LLM reasoning to resolve, or asks for clarification | Current Domain Agent |
| 5 | **Incomprehensible input** | Clarify → Options → Re-route → Escalate (ladder) | Current Agent → Receptionist → Escalation |
| 6 | **Tool failure** | Agent explains the issue, offers alternatives (retry, human, self-service) | Current Domain Agent |
| 7 | **Member requests human** | Immediate route to Escalation Agent with full context | Escalation Agent |
| 8 | **Returning member** | Session loaded from DynamoDB, resume paused flow | Session Lambda → Domain Agent |
| 9 | **Multi-topic request** | Receptionist identifies primary topic, handles first, then asks about second | Receptionist (sequencing) |
