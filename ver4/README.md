# Conversational AI Platform — Final Architecture Documentation

Comprehensive architecture documentation for team review and discussion. This captures all architectural options explored, the recommended 7-layer design, and detailed conversation flow handling patterns.

## Context

- **Current state:** Platform uses Amazon Lex with a front Lambda orchestrating conversation flows. Team is skilled in Lex. Most intents are FAQ-like with predefined prompts and separate Lex intents.
- **Goal:** Enhance the platform by leveraging Amazon Bedrock (LLM) while maintaining compliance, reducing risk, and enabling rapid use case onboarding.
- **Constraints:** Financial/banking industry compliance, team Lex expertise, text-chat-first priority, cost sensitivity.

---

## Documents

| # | Document | What It Covers |
|---|----------|---------------|
| 1 | [Architecture Options](01-architecture-options.md) | Three architecture options compared side-by-side: Lex-Enhanced, Hybrid Orchestrator, Bedrock Agent-First |
| 2 | [Recommended Architecture — 7 Layers](02-recommended-architecture.md) | The recommended hybrid 7-layer architecture with detailed diagrams, each layer explained |
| 3 | [Conversation Flow Handling](03-conversation-flow-handling.md) | How the platform handles: intent switching mid-flow, incomprehensible utterances, slot failures, multi-intent requests — with detailed examples and sequence diagrams |

---

## Quick Summary of Recommendation

**Hybrid Orchestrator (Option B)** — Evolve the existing Lex platform incrementally. Keep Lex for NLU (team knows it), enhance the existing **Lambda Orchestrator** to be the central brain, and leverage **Bedrock on-demand** for understanding assistance, knowledge retrieval, and complex flows. This:

- ✅ Preserves team's Lex skills and existing investment
- ✅ Reduces LLM risk (deterministic paths where possible, LLM only when needed)
- ✅ Enables rapid use case onboarding via config-driven orchestrator
- ✅ Addresses conversation recovery (intent switching, fallback resolution)
- ✅ Allows incremental Bedrock adoption (start small, expand with confidence)
