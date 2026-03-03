# LLM Co-Pilot Architecture — Brainstorm

![LLM Co-Pilot Architecture with Escalation Ladder](images/copilot_architecture.png)

## The Paradigm Shift

**Old thinking:** Three separate paths — deterministic, knowledge, agent — picked at the start based on task complexity.

**New thinking:** The LLM is not a separate path. It's an **always-available co-pilot** that any part of the system can invoke when it needs help understanding the user.

```
┌────────────────────────────────────────────────────────────┐
│  "The LLM is not a destination. It's a capability."        │
│                                                            │
│  Any turn, any slot, any point in the conversation —       │
│  if the system can't understand the user, it asks the LLM. │
│  The LLM understands, and hands control back.              │
└────────────────────────────────────────────────────────────┘
```

---

## The Problem We're Solving

**Scenario:** Deterministic path — checking account balance.

```
Lex:   "Which account would you like to check?"  (eliciting AccountType slot)
User:  "The one I usually pay my electric bill from"
Lex:   ??? (can't map this to a slot value like "checking" or "savings")
```

**Old approach:** Lex fails → FallbackIntent → conversation breaks or fully hands off to LLM Agent (overkill).

**New approach:** Lex fails to elicit the slot → **Dialog Lambda invokes Bedrock** → Bedrock understands "the bill-paying account = checking account" → Lambda returns the resolved slot to Lex → Deterministic flow continues.

The user never notices. The conversation stays in the deterministic flow. The LLM was a silent co-pilot for one turn.

---

## Architecture: LLM as a Turn-Level Co-Pilot

```mermaid
flowchart TD
    U["User Utterance"] --> Lex["Lex V2\n+ Assisted NLU"]
    
    Lex --> Check{"Can Lex\nunderstand?"}
    
    Check -->|"✅ Yes"| Flow["Continue\nNormal Flow"]
    Check -->|"❌ No — low confidence,\nambiguous slot,\nunresolved elicitation"| Hook["Dialog Code Hook\n(Lambda)"]
    
    Hook --> Context["Build LLM Context:\n• Conversation history\n• Current intent\n• What we're trying to resolve\n• Available slot values\n• Business rules"]
    
    Context --> Bedrock["Invoke Bedrock\n(Understanding Request)"]
    
    Bedrock --> Resolution{"What did\nLLM resolve?"}
    
    Resolution -->|"Slot value"| SetSlot["Set slot value\n→ Return to Lex"]
    Resolution -->|"Intent clarification"| Reclassify["Suggest correct intent\n→ Re-route in Lex"]
    Resolution -->|"Need more info"| Clarify["Generate clarifying\nquestion → Ask user"]
    Resolution -->|"Too complex for\nslot-based flow"| Escalate["Hand off to\nBedrock Agent\n(full agent mode)"]
    
    SetSlot --> Flow
    Reclassify --> Flow
    Clarify --> U
    Escalate --> Agent["Bedrock Agent\n(takes full control)"]

    style Bedrock fill:#e67e22,color:#fff
    style Flow fill:#27ae60,color:#fff
    style Agent fill:#e74c3c,color:#fff
```

---

## When Does the Co-Pilot Activate?

The co-pilot isn't about simple vs complex. It activates based on **understanding failures at specific points:**

| Trigger Point | What Fails | What LLM Does | Example |
|--------------|-----------|---------------|---------|
| **Intent Classification** | Lex can't match to any intent (low confidence) | Understands utterance semantically, suggests correct intent | "I think there's something wrong with my charges" → DisputeCharge intent |
| **Slot Elicitation** | User gives an ambiguous or indirect answer to a slot prompt | Resolves the natural language to a valid slot value | "The one I pay bills from" → AccountType = "checking" |
| **Slot Validation** | User provides a value that doesn't match expected format/type | Interprets and normalizes the value | "Twenty-five hundred" → Amount = 2500 |
| **Disambiguation** | Multiple intents could match | Asks a smart clarifying question | "Is this about your homeowner's or auto policy?" |
| **Mid-flow Confusion** | User changes topic or asks a side question mid-flow | Decides: answer the side question and resume, or re-route | User mid-claim: "Wait, am I even covered for this?" |
| **Context Carryover** | User refers to something mentioned earlier | Resolves references from conversation history | "Use the same address" → pulls address from earlier turn |

---

## The Dialog Lambda — The Bridge

The **Dialog Code Hook Lambda** is the key piece. It sits between Lex and Bedrock and acts as the bridge:

```mermaid
sequenceDiagram
    participant U as User
    participant Lex as Lex V2
    participant DL as Dialog Lambda<br/>(Code Hook)
    participant BR as Bedrock<br/>(Co-Pilot)
    participant FL as Fulfillment<br/>Lambda

    Note over Lex, DL: Deterministic flow — checking balance

    U->>Lex: "Check my balance"
    Lex->>Lex: ✅ Intent: CheckBalance (high confidence)
    Lex->>U: "Which account?"
    
    U->>Lex: "The one I pay my mortgage from"
    Lex->>DL: ❌ Can't resolve AccountType slot
    
    Note over DL: Co-pilot activation!
    
    DL->>BR: "User said 'the one I pay my mortgage from'.<br/>We need AccountType slot.<br/>Valid values: checking, savings, investment.<br/>Conversation context: [...]"
    
    BR->>DL: "AccountType = 'checking'<br/>Reasoning: mortgage payments typically<br/>come from checking accounts"
    
    DL->>Lex: Set AccountType = "checking"
    Lex->>FL: All slots filled → Fulfill
    FL->>Lex: Balance: $4,250.00
    Lex->>U: "Your checking account balance is $4,250.00"
    
    Note over U, FL: User never knew LLM was involved!
```

### What the Dialog Lambda Sends to Bedrock

```json
{
  "request_type": "slot_resolution",
  "context": {
    "intent": "CheckBalance",
    "current_slot": "AccountType",
    "slot_prompt": "Which account would you like to check?",
    "valid_values": ["checking", "savings", "investment", "credit_card"],
    "user_response": "The one I pay my mortgage from",
    "conversation_history": [
      {"role": "user", "text": "Check my balance"},
      {"role": "bot", "text": "Which account would you like to check?"},
      {"role": "user", "text": "The one I pay my mortgage from"}
    ],
    "customer_context": {
      "accounts": ["checking", "savings"]
    }
  },
  "instruction": "Resolve the user's response to one of the valid slot values. If you cannot determine the value with confidence, suggest a clarifying question."
}
```

### What Bedrock Returns

```json
{
  "resolved": true,
  "slot_value": "checking",
  "confidence": 0.92,
  "reasoning": "Mortgage payments are typically made from checking accounts. Customer has a checking account.",
  "clarifying_question": null
}
```

Or when it can't resolve:

```json
{
  "resolved": false,
  "slot_value": null,
  "confidence": 0.45,
  "reasoning": "User said 'the usual one' but has both checking and savings with similar activity",
  "clarifying_question": "I see you have both a checking and savings account. Which one would you like to check?"
}
```

---

## Escalation Ladder — When Does Co-Pilot Become Full Agent?

The co-pilot doesn't always resolve things in one turn. There's an **escalation ladder:**

```mermaid
flowchart TD
    L1["Level 1: Assisted NLU\n(Built into Lex V2)"] -->|"Still can't classify"| L2
    L2["Level 2: Dialog Lambda\nInvokes Bedrock for\nslot/intent resolution"] -->|"Still can't resolve\nafter 2 attempts"| L3
    L3["Level 3: LLM generates\nsmart clarifying question"] -->|"User response still\nambiguous or topic shift"| L4
    L4["Level 4: Hand off to\nBedrock Agent\n(full agent control)"]
    
    L1 -->|"✅ Resolved"| OK["Continue Flow"]
    L2 -->|"✅ Resolved"| OK
    L3 -->|"✅ Resolved"| OK

    style L1 fill:#27ae60,color:#fff
    style L2 fill:#2980b9,color:#fff
    style L3 fill:#e67e22,color:#fff
    style L4 fill:#e74c3c,color:#fff
    style OK fill:#27ae60,color:#fff
```

| Level | What Happens | Cost | Latency |
|-------|-------------|------|---------|
| **L1** | Lex Assisted NLU (automatic, built-in) | Low | ~50ms |
| **L2** | Dialog Lambda → Bedrock (single-turn understanding) | Medium | ~1-2s |
| **L3** | LLM generates clarifying question, waits for user | Medium | ~1-2s |
| **L4** | Full handoff to Bedrock Agent (takes over conversation) | Higher | ~2-5s |

**Key rule:** The system **tries to stay in the deterministic flow** as long as possible. It only escalates to full agent mode when the conversation has become too complex for turn-level LLM assistance.

---

## Revised Architecture Diagram

```mermaid
graph TB
    subgraph Channels["Channels"]
        Voice["📞 Amazon Connect"]
        Chat["💬 Web/Mobile Chat"]
        SMS["📱 SMS"]
    end

    subgraph Orchestrator["Lex V2 — Conversation Orchestrator"]
        ANLU["Assisted NLU\n(LLM-enhanced classification)"]
        Intents["Intent Definitions\n(Standard + QnA + BedrockAgent)"]
        Dialog["Dialog Management\n(Slot elicitation, confirmation)"]
    end

    subgraph CoPilot["🤖 LLM Co-Pilot (Bedrock)"]
        DLambda["Dialog Code Hook Lambda"]
        Resolve["Turn-Level Understanding:\n• Slot resolution\n• Intent clarification\n• Disambiguation\n• Context carryover\n• Smart clarifying Qs"]
    end

    subgraph Knowledge["Knowledge Layer"]
        KB["Bedrock Knowledge Bases\n(QnAIntent → RAG)"]
        OS["OpenSearch Serverless"]
        S3["S3 Documents"]
    end

    subgraph Agent["Full Agent Mode"]
        BA["Bedrock Agent\n(BedrockAgentIntent)"]
        AG["Action Groups\n(Tool Calling)"]
        SF["Step Functions\n(Multi-step Flows)"]
    end

    subgraph Fulfill["Deterministic Fulfillment"]
        FL["Fulfillment Lambdas"]
        API["Backend APIs"]
    end

    subgraph Guard["Safety Layer"]
        GR["Bedrock Guardrails\n(PII, Toxicity, Grounding, Denied Topics)"]
    end

    Channels --> Orchestrator
    ANLU -.->|"L1: Built-in\nLLM assist"| Intents
    Dialog -->|"❌ Can't understand"| DLambda
    DLambda --> Resolve
    Resolve -->|"✅ Resolved"| Dialog
    Resolve -->|"❌ Escalate"| BA
    
    Intents -->|"QnAIntent"| KB
    KB --> OS & S3
    Intents -->|"BedrockAgentIntent"| BA
    BA --> AG --> SF
    Intents -->|"Standard Intent"| FL --> API
    
    BA --> GR
    DLambda --> GR
    KB --> GR
```

---

## How This Changes Use Case Onboarding

The YAML spec now needs to define **co-pilot behavior** per intent:

```yaml
use_case:
  name: "check-balance"
  lob: "banking"
  
  intents:
    - name: CheckBalance
      type: standard           # deterministic flow
      
      slots:
        - name: AccountType
          type: AMAZON.AlphaNumeric
          prompt: "Which account would you like to check?"
          valid_values: ["checking", "savings", "investment", "credit_card"]
          
          # Co-pilot config for this slot
          copilot:
            enabled: true
            resolution_hint: "Resolve casual references to account types. 'Bill-paying' = checking, 'rainy day' = savings."
            max_attempts: 2        # escalate to agent after 2 failed co-pilot resolutions
            
      copilot:
        enabled: true
        escalation_threshold: 2   # after 2 co-pilot invocations per conversation, consider full agent handoff
        
      fulfillment:
        type: lambda
        handler: banking/check_balance

    - name: FileClaimComplex
      type: bedrock_agent         # starts in full agent mode
      agent_id: "insurance-claims-agent"
      # No copilot config needed — Bedrock Agent handles everything
```

---

## Key Design Questions to Resolve

1. **Cost control** — Every co-pilot invocation costs Bedrock tokens. Should we limit co-pilot calls per conversation (e.g., max 3)?

2. **Latency budget** — Co-pilot adds ~1-2s per invocation. For voice (real-time), is this acceptable? Should voice have a stricter escalation-to-agent threshold?

3. **Context window** — How much conversation history do we pass to the co-pilot? Full history or last N turns?

4. **Learning loop** — When the co-pilot resolves a slot successfully, should we feed that back as training data for Lex (so next time Lex handles it natively)?

5. **Observability** — Should co-pilot invocations be tracked as a separate metric tier? (How often is the co-pilot saving conversations vs. just adding latency?)
