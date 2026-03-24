# Multi-Agent Conversational AI Platform — Final Architecture

A strategic, multi-agent conversational platform for a financial institution offering banking and insurance products. This architecture replaces the Lex-centric hybrid model with a fully agentic, hierarchical structure powered by Amazon Bedrock Agents.

## Context

- **Previous state:** Hybrid Orchestrator architecture (ver4) using Lex V2 for NLU with a Lambda Orchestrator and on-demand Bedrock.
- **Evolution:** Lex is removed from the architecture entirely. The platform is now a hierarchy of specialized Bedrock Agents, each purpose-built for a domain, topic, or task.
- **Industry:** Financial services — banking (accounts, cards, disputes, transfers) + insurance (claims, quotes, policies).

---

## Documents

| # | Document | What It Covers |
|---|----------|----------------|
| 1 | [Vision & Principles](01-vision-and-principles.md) | What multi-agent conversation is, design principles, why we moved past Lex |
| 2 | [Architecture Overview](02-architecture-overview.md) | High-level architecture, agent hierarchy, data flow, AWS services |
| 3 | [Agent Definitions](03-agent-definitions.md) | Each agent's role, system instructions, tools, context contracts |
| 4 | [Orchestration & Recovery](04-orchestration-and-recovery.md) | Agent handoffs, intent switching, error recovery, escalation patterns |
| 5 | [Onboarding & Scaling](05-onboarding-and-scaling.md) | Adding new intents/agents, right-sizing strategy, future splitting |

---

## Architecture at a Glance

```
Member ──→ API Gateway ──→ Receptionist Agent (Router)
                              ├──→ Banking Agent ──→ Tools (Balance, Cards, Disputes, Transfers)
                              ├──→ Insurance Agent ──→ Tools (Claims, Quotes, Policies)
                              ├──→ Knowledge Agent ──→ Bedrock Knowledge Bases (RAG)
                              └──→ Escalation Agent ──→ Live Agent Handoff
```

## Diagrams

Editable architecture diagrams are in the [`diagrams/`](diagrams/) folder:
- `architecture-overview.excalidraw` — open with [excalidraw.com](https://excalidraw.com) for editing
- `architecture-overview.png` — rendered export for embedding in presentations
