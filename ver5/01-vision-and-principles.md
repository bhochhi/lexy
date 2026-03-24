# Vision & Design Principles

## Vision Statement

> **A multi-agent conversational platform is a system of specialized AI agents, organized in a hierarchy, that collaborate to understand, route, and resolve member inquiries — each agent doing one thing exceptionally well.**
>
> The platform operates like a well-run customer service organization. A **Receptionist** greets every member and understands *why* they're reaching out. Based on the topic, the Receptionist routes the conversation to the right **Domain Agent** — a specialist who knows everything about banking, insurance, or another product line. That specialist collects the information needed to resolve the inquiry, invokes the right back-end tools, and delivers a clear answer. If the member changes topics mid-conversation, the system gracefully transitions them to the right specialist without losing context.
>
> Some inquiries are **zero-shot** — the member asks a question, and the agent answers immediately from a knowledge base. Others are **multi-turn** — the agent needs to collect additional information (account number, claim details, date of birth) before it can complete the action. The architecture handles both seamlessly.
>
> Every agent is designed with **right-sized responsibility**: broad enough to handle its domain efficiently, narrow enough to remain fast, accurate, and maintainable. As the platform grows, any agent can be decomposed into sub-agents without redesigning the system.

---

## Why Move Beyond Lex?

Amazon Lex V2 served the team well for deterministic, slot-based conversations. But as the platform scales, Lex's limitations become blocking:

### Lex Limitations That Drive This Change

| Limitation | Impact |
|-----------|--------|
| **Assisted NLU slot types are restricted** | Lex Assisted NLU only works with Lex's built-in slot types. Custom slot types with complex validation aren't supported, limiting the kinds of information agents can collect. |
| **Generated responses are restricted** | Lex constrains how responses are generated — you can't freely compose dynamic, context-aware responses. In a financial context, this limits the ability to provide personalized, nuanced answers. |
| **Rigid intent/slot model** | Every conversational path must be pre-defined as an intent with slots. Natural, free-form conversations don't fit this model. |
| **Mid-flow flexibility is limited** | Intent switching, multi-intent handling, and graceful recovery require complex orchestrator workarounds outside of Lex. |
| **NLU is a classification bottleneck** | Lex classifies one intent per turn. Real conversations are messier — members ask multiple things, provide context narratively, or switch topics. |
| **Training overhead** | Each new intent requires utterance samples, slot definitions, and bot rebuilds. This slows onboarding. |

### What Multi-Agent Gives Us

| Capability | How Multi-Agent Solves It |
|-----------|--------------------------|
| **Natural language flexibility** | Each agent uses a Bedrock LLM — no pre-defined utterances needed. The agent understands intent from natural language. |
| **Dynamic information collection** | Agents collect information conversationally, not through rigid slot elicitation. They can ask follow-up questions naturally. |
| **Graceful transitions** | Agent-to-agent handoffs with full context passing. The member never feels "lost" in the system. |
| **Zero-to-multi-turn** | An agent can answer a FAQ instantly (0-shot) or conduct a multi-turn dialogue to collect information — same architecture, same agent. |
| **Rapid onboarding** | New capability = new tool definition on an existing agent, or a new agent with its own instructions. No Lex training required. |
| **Right-sized specialization** | Each agent has focused instructions and tools. This produces better responses than a monolithic bot trying to do everything. |

---

## Design Principles

### 1. Specialization Over Generalization

Each agent has a **clear, bounded responsibility**. The Receptionist doesn't try to resolve banking questions — it identifies the topic and routes. The Banking Agent doesn't know about insurance — it handles banking and nothing else. This keeps instructions focused, reduces hallucination, and makes each agent easier to test and maintain.

### 2. Right-Sized Agents, Ready to Split

An agent should handle a **coherent domain** — not too broad (slow, unfocused) and not too narrow (too many agents, overhead). Start with domain-level agents (Banking, Insurance). When an agent grows beyond ~15 tools or its instruction set exceeds what the model can reliably follow, split it into sub-agents (Banking → Accounts Agent, Cards Agent, Disputes Agent).

### 3. Deterministic Where Possible, LLM Where Needed

Even in a multi-agent architecture, not every response needs LLM generation. Simple lookups (balance check, claim status) should use tools that return structured data. The agent frames the response, but the *data* comes from deterministic APIs. This reduces cost, latency, and compliance risk.

### 4. Graceful Recovery Over Hard Failure

When a member says something unexpected, the agent doesn't fail — it **recovers**:
1. **Clarify** — ask the member to rephrase
2. **Re-route** — if the utterance belongs to another agent, hand off with context
3. **Escalate** — if the agent can't resolve after multiple attempts, transfer to a human with full conversation history

### 5. Context Flows Downstream, Results Flow Upstream

The Receptionist passes **context** (member ID, identified topic, initial utterance) downstream to domain agents. Domain agents pass **results** (action taken, data retrieved, resolution status) upstream. No agent needs to understand the full system — just its inputs and outputs.

### 6. Guardrails Are Non-Negotiable

Every agent's output passes through Bedrock Guardrails. In financial services, this means:
- PII redaction on all responses
- Denied topics (investment advice, guaranteed returns)
- Grounding checks (responses must be based on actual data, not hallucinated)
- Toxicity filtering

### 7. Observable by Default

Every agent interaction is traced: which agent handled the turn, what tools were invoked, how long it took, whether the member's issue was resolved. This data feeds a continuous improvement loop — identify underperforming agents, refine their instructions, expand their tools.

---

## The Agent Hierarchy — Mental Model

Think of the platform as a financial services call center:

```
┌─────────────────────────────────────────────────────────┐
│                    RECEPTIONIST                         │
│  "Welcome! Are you calling about banking, insurance,    │
│   or something else?"                                   │
│                                                         │
│  Knows: All topics at a high level                      │
│  Does: Identifies topic, routes to specialist           │
│  Doesn't: Answer detailed questions or take actions     │
└──────────┬──────────────┬──────────────┬────────────────┘
           │              │              │
    ┌──────▼──────┐ ┌─────▼──────┐ ┌────▼─────────┐
    │  BANKING    │ │ INSURANCE  │ │  KNOWLEDGE   │
    │  SPECIALIST │ │ SPECIALIST │ │  DESK        │
    │             │ │            │ │              │
    │ Knows:      │ │ Knows:     │ │ Knows:       │
    │ Accounts,   │ │ Claims,    │ │ General FAQs,│
    │ Cards,      │ │ Quotes,    │ │ Product info,│
    │ Disputes,   │ │ Policies,  │ │ Procedures   │
    │ Transfers   │ │ Renewals   │ │              │
    └─────────────┘ └────────────┘ └──────────────┘
```

Each specialist can, in the future, be broken down further:

```
  BANKING SPECIALIST (today)
  ├── Accounts (balance, statements, open/close)
  ├── Cards (activate, block, replace, limits)
  ├── Disputes (file, track, resolve)
  └── Transfers (internal, external, wire)

  ↓ When banking grows to 20+ tools...

  BANKING ROUTER (future)
  ├── Accounts Agent
  ├── Cards Agent
  ├── Disputes Agent
  └── Transfers Agent
```

This is the power of the hierarchical model — **the architecture scales by splitting, not by rewriting.**

---

## How a Conversation Works (Conceptual)

### Zero-Shot (FAQ)

```
Member: "What are your branch hours?"
  → Receptionist: This is a general knowledge question → route to Knowledge Agent
  → Knowledge Agent: Retrieves from Knowledge Base → "Our branches are open Mon-Fri, 9am-5pm..."
  → Done. One turn. No multi-turn needed.
```

### Multi-Turn (Action)

```
Member: "I need to dispute a charge"
  → Receptionist: This is banking-related → route to Banking Agent
  → Banking Agent: "I can help with that. Which transaction would you like to dispute?"
  → Member: "The $47.99 from yesterday"
  → Banking Agent: Calls tool: getRecentTransactions(memberId) → finds the $47.99 charge
  → Banking Agent: "I found a $47.99 charge at Coffee Shop on March 10. Why are you disputing this?"
  → Member: "I didn't make that purchase"
  → Banking Agent: Calls tool: fileDispute(transactionId, reason="unauthorized")
  → Banking Agent: "I've filed dispute #DSP-1234. You'll receive a provisional credit within 48 hours."
  → Done. Multi-turn. Agent collected what it needed conversationally.
```

### Mid-Conversation Topic Switch

```
Member: "I need to dispute a charge"
  → Receptionist → Banking Agent
  → Banking Agent: "Which transaction?"
  → Member: "Wait, first — did my home insurance claim get approved?"
  → Banking Agent: This is insurance, not banking. → Signals Receptionist to re-route.
  → Receptionist: Routes to Insurance Agent with context (member was in banking, paused flow)
  → Insurance Agent: Answers the claim question.
  → Receptionist: "Now, would you like to continue with the dispute?"
  → Member: "Yes"
  → Banking Agent: Resumes from where they left off.
```
