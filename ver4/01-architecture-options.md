# Architecture Options — Side-by-Side Comparison

Three architecture options explored during our design process, evaluated against the team's real-world constraints.

---

## Our Constraints (Evaluation Criteria)

| # | Constraint | Weight | Notes |
|---|-----------|--------|-------|
| 1 | **Text-chat first** | Critical | Primary channel; voice is secondary |
| 2 | **Compliance & risk** | Critical | Financial/banking — must reduce LLM liability |
| 3 | **Team skills** | High | Team knows Lex, limited LLM/Bedrock experience |
| 4 | **Onboarding speed** | High | New intent = POC in days, prod in weeks |
| 5 | **Promotion simplicity** | High | POC → Staging → Prod must be simple |
| 6 | **Cost** | Medium | Per-conversation cost matters at scale |
| 7 | **Observability** | Medium | Per-intent metrics, quality tracking |
| 8 | **Scalability** | Medium | Support many LOBs, many intents |

---

## Option A: Lex-Enhanced (Evolutionary)

> *"Upgrade Lex with its new LLM features, keep everything else the same."*

```
Text/Voice → Lex V2 (Assisted NLU) → Lambda Fulfillment
                                    ↘ QnAIntent → Bedrock KB
                                    ↘ BedrockAgentIntent → Bedrock Agent
```

### How It Works
- Keep existing Lex bot structure and intents
- Enable **Assisted NLU** (LLM-enhanced classification) for better accuracy
- Add **QnAIntent** for FAQ responses via Bedrock Knowledge Bases
- Add **BedrockAgentIntent** for complex intents that need LLM reasoning
- Keep existing front Lambda for orchestration

### Pros
- ✅ Minimal change from current architecture
- ✅ Team keeps using Lex — low learning curve
- ✅ Bedrock features are "opt-in" per intent
- ✅ Lower risk — existing deterministic flows unchanged

### Cons
- ❌ Still requires Lex intent + utterance training for each new intent
- ❌ Lex is still the bottleneck for understanding
- ❌ Limited control over how Assisted NLU/QnAIntent behaves
- ❌ Text channel still routes through Lex (unnecessary for text)
- ❌ Mid-flow intent switching and recovery are still hard

### Onboarding a New Intent
1. Define Lex intent with sample utterances
2. Configure slots and prompts
3. Write fulfillment Lambda
4. Train and build Lex bot version
5. Test and deploy

**Effort: ~3-5 days per intent**

---

## Option B: Hybrid Orchestrator (Recommended)

> *"Keep Lex for NLU, but make the Lambda Orchestrator the smart brain. Bedrock assists on-demand."*

```
Text/Voice → Lex V2 (NLU) → Orchestrator Lambda (Central Brain)
                                ├→ Deterministic fulfillment
                                ├→ Bedrock (understanding assistance)
                                ├→ Bedrock KB (knowledge retrieval)
                                ├→ Bedrock Agent (complex flow takeover)
                                └→ Recovery / re-routing logic
```

### How It Works
- Lex classifies intents and elicits slots (what the team knows)
- The **Orchestrator Lambda** becomes the central brain — it decides:
  - Can I fulfill this deterministically? → Do it (fast, cheap)
  - Do I need knowledge? → Query Bedrock KB
  - Can't understand the user? → Ask Bedrock to help resolve
  - Is this conversation too complex? → Delegate to Bedrock Agent
  - Did the user switch intents? → Handle the transition gracefully
- **Bedrock is "on-demand"** — the orchestrator invokes it only when needed

### Pros
- ✅ Preserves Lex investment and team skills
- ✅ Orchestrator has full control — deterministic by default, LLM when needed
- ✅ Handles intent switching, fallback recovery, re-routing (in orchestrator code)
- ✅ New intents can be config-driven (orchestrator reads intent config from DynamoDB/YAML)
- ✅ Compliance-friendly — LLM is behind guardrails and only invoked when the orchestrator decides
- ✅ Incremental adoption path — start with Lex-only, add Bedrock features gradually

### Cons
- ⚠️ Orchestrator Lambda is now a critical, complex component — needs good design
- ⚠️ More code in the orchestrator than a pure Lex or pure Agent approach
- ⚠️ Two systems to maintain (Lex + Bedrock) vs. one

### Onboarding a New Intent
1. Add intent config to the registry (YAML/DynamoDB)
2. Define Lex intent (or reuse BedrockAgentIntent for LLM-driven intents)
3. Write fulfillment Lambda (or tool definition if using Bedrock Agent)
4. Deploy via GitLab pipeline

**Effort: ~1-3 days per intent (config-driven)**

---

## Option C: Bedrock Agent-First (Modern)

> *"LLM is the brain. No Lex for text. Lex only for voice ASR."*

```
Text → API Gateway → Bedrock Agent (Brain) → Tools (Lambdas)
Voice → Connect → Lex (ASR only) → Same Bedrock Agent
```

### How It Works
- For text: no Lex at all — Bedrock Agent handles everything
- Each "intent" is a **tool definition** (JSON schema) in the Agent's Action Group
- The Agent decides which tool to call based on the user's natural language
- For voice: Lex provides ASR (speech-to-text) and immediately delegates to the Agent via BedrockAgentIntent

### Pros
- ✅ Simplest architecture — fewest components
- ✅ Adding a new intent = add a tool definition + Lambda (2 steps)
- ✅ Best natural language understanding — LLM handles everything
- ✅ Built-in disambiguation, clarification, multi-turn
- ✅ No Lex training or utterance management

### Cons
- ❌ **Every turn uses LLM tokens** — higher per-conversation cost
- ❌ **LLM compliance risk** — every response generated by LLM, even simple balance checks
- ❌ **Team must learn Bedrock Agent development** — significant skill gap
- ❌ Less control over exact conversation flow — LLM is probabilistic
- ❌ Guardrails become critical (team must learn to configure them)
- ❌ Debugging is harder — "why did the agent say that?" is harder than deterministic flows

### Onboarding a New Intent
1. Add tool definition (JSON) to Agent Action Group
2. Write Lambda handler

**Effort: ~1 day per intent (if team is skilled)**

---

## Head-to-Head Comparison

| Criteria | Option A: Lex-Enhanced | Option B: Hybrid Orchestrator | Option C: Agent-First |
|----------|----------------------|------------------------------|----------------------|
| **Text-first** | ⚠️ Lex in text path | ⚠️ Lex in text path* | ✅ No Lex for text |
| **Compliance / risk** | ✅ Mostly deterministic | ✅ Deterministic default, LLM opt-in | ❌ Full LLM for every turn |
| **Team skills** | ✅ Lex-native | ✅ Lex + incremental Bedrock | ❌ Requires LLM expertise |
| **Onboarding speed** | ⚠️ 3-5 days (Lex training) | ✅ 1-3 days (config-driven) | ✅ 1 day (tool definition) |
| **Promotion simplicity** | ⚠️ Lex bot versions + Lambda | ✅ Config + Lambda | ✅ Agent config + Lambda |
| **Cost per conversation** | ✅ Low (mostly no LLM) | ✅ Medium (LLM on-demand) | ⚠️ Higher (every turn uses LLM) |
| **Intent switching / recovery** | ❌ Manual, fragile | ✅ Orchestrator handles | ✅ Agent handles naturally |
| **Observability** | ⚠️ Lex built-in only | ✅ Full (orchestrator tracks everything) | ⚠️ Need custom metrics |
| **Control / predictability** | ✅ High (deterministic) | ✅ High (orchestrator decides) | ⚠️ Medium (LLM is probabilistic) |

*\* Lex is in the text path for NLU classification, but the orchestrator controls everything after that.*

---

## Risk Matrix

| Risk | Option A | Option B | Option C |
|------|---------|---------|---------|
| LLM generates unauthorized financial advice | Low | Low (guardrails + deterministic default) | **High** (every response is LLM-generated) |
| LLM hallucination | Low | Low (LLM invoked selectively) | **Medium** (mitigated by guardrails) |
| Team can't build/maintain | Low | **Medium** (orchestrator complexity) | **High** (full Bedrock expertise needed) |
| Cost escalation at scale | Low | Medium | **High** |
| Conversation quality | Medium (Lex limitations) | **High** (best of both) | High (LLM is good) |
| Vendor lock-in | High (Lex) | Medium (Lex + Bedrock) | Medium (Bedrock) |
