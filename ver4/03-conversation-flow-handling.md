# Conversation Flow Handling — Intent Switching, Recovery & Orchestrator Patterns

How the platform handles real-world conversation challenges: mid-flow intent switching, incomprehensible utterances, slot resolution failures, multi-intent requests, and graceful recovery. All examples use the Hybrid Orchestrator (Option B).

---

## The Orchestrator's Decision Loop

Every conversational turn flows through this loop:

```mermaid
flowchart TD
    Turn["User sends message"] --> Lex["Lex classifies\n(intent + slots)"]
    Lex --> Orch["Orchestrator receives:\n• intent (or FallbackIntent)\n• slots (filled/missing)\n• session state\n• conversation history"]
    
    Orch --> D1{"Intent\nclassified?"}
    D1 -->|"Yes"| D2{"Same intent\nas current flow?"}
    D1 -->|"No (FallbackIntent)"| Recovery["Recovery Engine\n(see Section 3)"]
    
    D2 -->|"Yes"| D3{"All required\nslots filled?"}
    D2 -->|"No — different intent"| Switch["Intent Switch Handler\n(see Section 2)"]
    
    D3 -->|"Yes"| Fulfill["Fulfill\n(deterministic or Bedrock)"]
    D3 -->|"No"| D4{"Lex couldn't\nelicit slot?"}
    
    D4 -->|"Normal elicitation"| Elicit["Return slot prompt\nto user"]
    D4 -->|"User response unclear"| Assist["Bedrock Understanding\nAssist"]
    
    Assist -->|"Resolved"| D3
    Assist -->|"Unresolved"| Clarify["Ask clarifying\nquestion"]
    
    Recovery -->|"Resolved"| D1
    Recovery -->|"Failed N times"| Escalate["Escalate to\nhuman agent"]
    
    Switch --> Orch

    style Orch fill:#2c3e50,color:#fff
    style Recovery fill:#e67e22,color:#fff
    style Assist fill:#3498db,color:#fff
    style Escalate fill:#e74c3c,color:#fff
```

---

## 1. Happy Path — Single Intent, Resolved Successfully

The simplest case. Lex classifies correctly, slots fill normally, fulfillment succeeds.

### Example: Check Account Balance

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant L as Lex V2
    participant O as Orchestrator
    participant F as Fulfillment Lambda
    participant API as Banking API

    U->>L: "What's my checking account balance?"
    Note over L: Intent: CheckBalance ✅<br/>Slot: accountType = "checking" ✅
    L->>O: {intent: "CheckBalance", slots: {accountType: "checking"}}
    O->>O: All slots filled → Deterministic fulfillment
    O->>F: Invoke check_balance_handler
    F->>API: getBalance(customerId, "checking")
    API-->>F: $4,250.00
    F-->>O: Response
    O-->>L: "Your checking account balance is $4,250.00"
    L-->>U: "Your checking account balance is $4,250.00"
```

**Cost:** $0 Bedrock usage. Pure Lex + Lambda. ~200ms latency.

---

## 2. Intent Switching Mid-Flow

**The challenge:** User starts one flow, then changes their mind or asks about something else mid-conversation.

### Example 2a: Clean Intent Switch

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant L as Lex V2
    participant O as Orchestrator
    participant DDB as DynamoDB

    U->>L: "I want to check my balance"
    Note over L: Intent: CheckBalance
    L->>O: {intent: "CheckBalance", slots: {}}
    O->>DDB: Save active flow: CheckBalance
    O-->>L: "Which account — checking or savings?"
    L-->>U: "Which account — checking or savings?"

    U->>L: "Actually, I want to dispute a charge"
    Note over L: Intent: DisputeCharge ⚠️ (different!)
    L->>O: {intent: "DisputeCharge", slots: {}}
    
    Note over O: 🧠 Intent Switch Detected!<br/>Previous: CheckBalance<br/>New: DisputeCharge
    
    O->>DDB: Pause CheckBalance flow (save state)
    O->>DDB: Start DisputeCharge flow
    O-->>L: "Sure, I can help with a dispute. Which transaction are you referring to?"
    L-->>U: "Sure, I can help with a dispute. Which transaction are you referring to?"
```

**What the Orchestrator does:**
1. Detects the intent change (current session was `CheckBalance`, new request is `DisputeCharge`)
2. Saves the paused flow state in DynamoDB (in case user wants to return)
3. Starts the new flow
4. Optionally: after the new flow completes, asks "Would you still like to check your balance?"

### Example 2b: Ambiguous Intent Switch (Lex Reclassifies vs. Slot Response)

This is the tricky case — the user says something during slot elicitation that *looks* like a new intent:

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant L as Lex V2
    participant O as Orchestrator
    participant BR as Bedrock<br/>(Understanding Assist)

    U->>L: "I want to file a claim"
    Note over L: Intent: FileClaim
    L->>O: {intent: "FileClaim", slots: {}}
    O-->>L: "What type of claim —  auto, home, or life?"
    L-->>U: "What type of claim — auto, home, or life?"

    U->>L: "I was in a car accident on the highway yesterday"
    
    Note over L: 🤔 Is this:<br/>A) Slot value for claimType = "auto"?<br/>B) New intent: ReportAccident?<br/>C) Just context/narrative?
    
    alt Lex classifies as new intent (ReportAccident)
        L->>O: {intent: "ReportAccident", slots: {}}
        Note over O: 🧠 Intent switch detected<br/>BUT user was in FileClaim flow<br/>and this is contextually related
        O->>O: Check: Is ReportAccident related to FileClaim?
        O->>O: Yes — "car accident" implies claimType = "auto"
        O->>O: Merge context: set claimType = "auto" + store accident details
        O-->>L: "I understand — a car accident. Let me start your auto claim. When exactly did this happen?"
    else Lex can't classify (FallbackIntent)
        L->>O: {intent: "FallbackIntent", activeFlow: "FileClaim"}
        Note over O: 🧠 User is in FileClaim flow<br/>Lex couldn't classify their response<br/>But we're eliciting claimType slot
        O->>BR: "User is filing a claim. Asked for claim type. User said: 'I was in a car accident on the highway yesterday.' What claim type is this?"
        BR-->>O: {value: "auto", reasoning: "car accident = auto claim"}
        O->>O: Set claimType = "auto"
        O-->>L: "Got it — an auto claim for a car accident. Can you provide the date of the incident?"
    end
```

**The orchestrator's logic:**
```python
def handle_turn(event, session):
    new_intent = event['intent']
    active_flow = session.get('active_flow')
    
    # Case 1: No active flow — just route normally
    if not active_flow:
        return route_intent(new_intent, event['slots'])
    
    # Case 2: Same intent — continue flow
    if new_intent == active_flow['intent']:
        return continue_flow(active_flow, event['slots'])
    
    # Case 3: FallbackIntent during active flow — user is probably
    # responding to our question, Lex just didn't understand
    if new_intent == 'FallbackIntent':
        return attempt_resolution(active_flow, event['user_text'])
    
    # Case 4: Different intent — is it related to active flow?
    if is_related_intent(new_intent, active_flow['intent']):
        return merge_into_flow(active_flow, new_intent, event)
    
    # Case 5: Completely different intent — clean switch
    return switch_intent(active_flow, new_intent, event)
```

---

## 3. Incomprehensible Utterances — Recovery Engine

**The challenge:** User says something Lex cannot classify at all.

### Recovery Strategy Ladder

```mermaid
flowchart TD
    Fail["Lex → FallbackIntent"] --> R1
    
    R1["Attempt 1: Bedrock Understanding Assist"] -->|"Resolved"| Success["Resume flow ✅"]
    R1 -->|"Unresolved"| R2
    
    R2["Attempt 2: Offer options/buttons"] -->|"User selects"| Success
    R2 -->|"User ignores / another unclear response"| R3
    
    R3["Attempt 3: Rephrase question + examples"] -->|"Resolved"| Success
    R3 -->|"Still unclear"| R4
    
    R4["Attempt 4: Escalate"] --> Escalate["Transfer to human agent\n(with full conversation context)"]

    style R1 fill:#3498db,color:#fff
    style R2 fill:#e67e22,color:#fff
    style R3 fill:#e67e22,color:#fff
    style R4 fill:#e74c3c,color:#fff
    style Success fill:#27ae60,color:#fff
```

### Example 3a: Nonsensical Input

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant L as Lex V2
    participant O as Orchestrator
    participant BR as Bedrock

    U->>L: "asdlkfj asdf"
    Note over L: FallbackIntent
    L->>O: {intent: "FallbackIntent"}
    O->>O: Attempt 1: Check if user is in active flow
    O->>O: No active flow — this is a fresh message
    O-->>L: "I didn't quite catch that. How can I help you today? For example, you can ask about your account balance, file a claim, or get a quote."
    L-->>U: "I didn't quite catch that. How can I help you today?"
    
    U->>L: "blah blah"
    L->>O: {intent: "FallbackIntent", attempt: 2}
    O->>O: Attempt 2: Offer structured options
    O-->>L: Display quick reply buttons: [Check Balance] [File Claim] [Get Quote] [Talk to Agent]
    L-->>U: Quick reply buttons displayed
    
    U->>L: Clicks [Check Balance]
    Note over L: Intent: CheckBalance ✅
    L->>O: {intent: "CheckBalance"}
    O-->>L: "Which account would you like to check?"
    Note over U, L: Conversation recovered! ✅
```

### Example 3b: Unclear Slot Response With Bedrock Assist

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant L as Lex V2
    participant O as Orchestrator
    participant BR as Bedrock

    U->>L: "Check my balance"
    Note over L: Intent: CheckBalance ✅
    L->>O: {intent: "CheckBalance", slots: {accountType: null}}
    O-->>L: "Which account — checking, savings, or investment?"
    L-->>U: "Which account — checking, savings, or investment?"

    U->>L: "the one where my paycheck goes"
    Note over L: Can't resolve slot<br/>accountType = null
    L->>O: {intent: "CheckBalance", slots: {accountType: null}, userText: "the one where my paycheck goes"}
    
    Note over O: 🧠 Slot unresolved — invoke Bedrock assist
    
    O->>BR: "User asked for account type. Valid: [checking, savings, investment]. User said: 'the one where my paycheck goes'. Customer has: checking, savings accounts. Resolve."
    BR-->>O: {value: "checking", confidence: 0.95, reasoning: "paychecks are direct-deposited to checking accounts"}
    
    Note over O: ✅ High confidence — accept resolution
    
    O->>O: Set accountType = "checking"
    O-->>L: "Your checking account balance is $4,250.00"
    L-->>U: "Your checking account balance is $4,250.00"
    
    Note over U, L: User never knew LLM was involved. ✅
```

### Example 3c: Low Confidence — Ask for Clarification

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant L as Lex V2
    participant O as Orchestrator
    participant BR as Bedrock

    U->>L: "Check my balance"
    L->>O: {intent: "CheckBalance", slots: {accountType: null}}
    O-->>L: "Which account?"
    L-->>U: "Which account?"

    U->>L: "the usual one"
    L->>O: {slots: {accountType: null}, userText: "the usual one"}
    
    O->>BR: "Resolve 'the usual one' to account type. Valid: [checking, savings, investment]."
    BR-->>O: {value: null, confidence: 0.30, reasoning: "ambiguous — 'usual' could be any account", clarifying_question: "Do you mean your checking or savings account?"}
    
    Note over O: ⚠️ Low confidence — use Bedrock's clarifying question
    
    O-->>L: "Do you mean your checking or savings account?"
    L-->>U: "Do you mean your checking or savings account?"
    
    U->>L: "Checking"
    Note over L: accountType = "checking" ✅
    L->>O: {slots: {accountType: "checking"}}
    O-->>L: "Your checking account balance is $4,250.00"
```

---

## 4. Mid-Flow Topic Detour (Side Question)

**The challenge:** User asks a related but off-topic question in the middle of a flow, then wants to resume.

### Example: Side Question During Claim Filing

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant O as Orchestrator
    participant BR as Bedrock KB

    U->>O: "I want to file a claim"
    O-->>U: "What type of claim?"
    U->>O: "Auto — I was in an accident"
    O->>O: claimType = "auto" ✅
    O-->>U: "When did the accident occur?"
    
    U->>O: "Wait — am I even covered for this? What does my policy cover?"
    
    Note over O: 🧠 Detour detected!<br/>User is asking a knowledge question<br/>but is IN a FileClaim flow
    
    O->>O: Pause FileClaim flow (save state)
    O->>BR: Query KB: "auto insurance coverage"
    BR-->>O: "Your auto policy covers collision, comprehensive, and liability..."
    
    O-->>U: "Your auto policy covers collision damage, comprehensive coverage, and liability. Based on your description, a collision claim would apply. ✅\n\nNow, back to your claim — when did the accident occur?"
    
    Note over U, O: Answered the side question AND resumed the flow ✅
```

**Orchestrator logic:**
```python
def handle_detour(active_flow, new_intent, user_text):
    """Handle a side question during an active flow."""
    
    # Check if it's a knowledge question
    if new_intent == 'QnAIntent' or is_knowledge_question(user_text):
        # Answer the question
        answer = query_knowledge_base(user_text)
        
        # Resume the active flow
        next_prompt = get_next_slot_prompt(active_flow)
        
        return f"{answer}\n\nNow, back to your {active_flow['intent']} — {next_prompt}"
    
    # If it's a completely different intent, do a clean switch
    return switch_intent(active_flow, new_intent)
```

---

## 5. Multi-Intent Request (User Asks Two Things at Once)

### Example: "Check my balance and also dispute a charge"

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant L as Lex V2
    participant O as Orchestrator
    participant BR as Bedrock

    U->>L: "Check my balance and also I need to dispute a charge"
    
    Note over L: Lex can only classify ONE intent<br/>Classifies as: CheckBalance<br/>Misses: DisputeCharge
    
    L->>O: {intent: "CheckBalance"}
    
    Note over O: 🧠 Orchestrator checks:<br/>Is the user text longer/more complex<br/>than expected for CheckBalance?
    
    O->>BR: "User said: 'Check my balance and also I need to dispute a charge.' Detected intents: CheckBalance. Are there additional intents in this message? List them."
    BR-->>O: {intents: ["CheckBalance", "DisputeCharge"]}
    
    Note over O: Multi-intent detected!<br/>Queue: [CheckBalance, DisputeCharge]
    
    O->>O: Process CheckBalance first (simpler)
    O-->>L: "Let me help with both. First, your checking account balance is $4,250.00. ✅\n\nNow, about the dispute — which transaction are you referring to?"
    
    Note over O: Active flow: DisputeCharge<br/>Queued: none ✅
```

---

## 6. Bedrock Agent Takeover and Handback

When the orchestrator delegates to a Bedrock Agent for complex flows, the Agent manages the conversation independently until resolution, then returns control.

```mermaid
sequenceDiagram
    participant U as 👤 User
    participant O as Orchestrator
    participant A as Bedrock Agent<br/>(Claims)
    participant T as Tools<br/>(Action Groups)

    U->>O: "I need to file a claim — I was in an accident"
    
    Note over O: Intent: FileClaim<br/>Config type: bedrock_agent<br/>→ Delegate to Claims Agent
    
    O->>A: Invoke Agent with conversation context
    
    loop Agent manages conversation
        A->>U: "I'm sorry to hear that. Can you describe what happened?"
        U->>A: "A car rear-ended me at a red light"
        A->>T: checkCoverage(policyId, "collision")
        T-->>A: Coverage confirmed, $500 deductible
        A->>U: "You're covered for collision with a $500 deductible. I'll need a few details to file the claim..."
        U->>A: Provides details
        A->>T: fileClaim(details)
        T-->>A: Claim CLM-789 created
    end
    
    A->>O: Agent complete. Claim filed: CLM-789
    
    Note over O: Agent returned control<br/>Resume orchestrator
    
    O-->>U: "Your claim CLM-789 has been filed. You'll receive updates via email. Is there anything else I can help with?"
```

---

## 7. Orchestrator State Machine — Complete View

The orchestrator tracks conversation state for each session:

```json
{
  "sessionId": "sess-12345",
  "customerId": "cust-67890",
  "active_flow": {
    "intent": "CheckBalance",
    "state": "eliciting_slot",
    "current_slot": "accountType",
    "slots_filled": {},
    "attempt_count": 1,
    "bedrock_assist_count": 0
  },
  "paused_flows": [],
  "queued_intents": [],
  "conversation_history": [
    {"role": "user", "text": "Check my balance", "timestamp": "..."},
    {"role": "bot", "text": "Which account?", "timestamp": "..."}
  ],
  "metrics": {
    "total_turns": 2,
    "bedrock_invocations": 0,
    "intent_switches": 0,
    "recovery_attempts": 0
  }
}
```

---

## 8. Recovery Configuration Per Intent

Each intent in the registry can define its own recovery behavior:

```yaml
intents:
  CheckBalance:
    recovery:
      slot_resolution:
        accountType:
          bedrock_assist: true
          fallback_options: ["checking", "savings", "investment"]  # show buttons
          max_bedrock_attempts: 2
      on_fallback_intent:
        strategy: bedrock_resolve    # ask Bedrock to understand
        max_attempts: 2
        then: offer_buttons          # if Bedrock can't resolve, show options
      on_repeated_failure:
        action: escalate
        message: "Let me connect you with a team member who can help."

  DisputeCharge:
    recovery:
      on_fallback_intent:
        strategy: bedrock_agent      # this is complex — hand to agent immediately
      on_repeated_failure:
        action: escalate
        priority: high               # disputes get priority escalation
```

---

## Summary: How the Orchestrator Handles Every Scenario

| Scenario | Orchestrator Action |
|----------|-------------------|
| **Happy path** | Lex classifies → Orchestrator fulfills deterministically |
| **Lex classifies with low confidence** | Assisted NLU (L1) kicks in automatically |
| **Slot can't be resolved** | Orchestrator invokes Bedrock to understand → resolved or clarifying question |
| **FallbackIntent (no active flow)** | Retry with Bedrock understanding → offer buttons → escalate |
| **FallbackIntent (during active flow)** | Assume user is responding to current prompt → Bedrock resolves in context |
| **Intent switch (clean)** | Pause current flow → start new flow → optionally return |
| **Intent switch (related)** | Merge into current flow (e.g., accident context into claim) |
| **Side question / detour** | Answer from KB → resume active flow |
| **Multi-intent** | Bedrock extracts all intents → queue and process sequentially |
| **Complex flow** | Delegate to Bedrock Agent → Agent returns control when done |
| **Repeated failures** | Escalate to human agent with full conversation context |
