# Conversational AI Platform — Architecture Documentation

Architecture documentation for a multi-channel, AWS-native Conversational AI platform serving insurance, banking, and advisory products. Uses a **hybrid Lex V2 + Bedrock** architecture with three routing paths: deterministic, knowledge, and LLM agent.

---

## 📄 Documents

| # | Document | Description |
|---|----------|-------------|
| 1 | [System Architecture](docs/01-system-architecture.md) | 7-layer platform architecture — channels, NLU, Bedrock LLM orchestration, knowledge & data, backend, observability, and DevOps |
| 2 | [Onboarding Pipeline](docs/02-onboarding-pipeline.md) | Config-driven use case onboarding, GitLab CI/CD pipeline, Terraform + CDK IaC strategy, YAML spec format, and POC-to-Prod promotion |
| 3 | [Observability](docs/03-observability.md) | 3-tier observability: infrastructure health (CloudWatch/X-Ray), conversation analytics (Athena/QuickSight), and LLM-powered quality evaluation |
| 4 | [Security & Compliance](docs/04-security-compliance.md) | Security architecture, PII handling, encryption, IAM strategy, and compliance frameworks (SOC 2, PCI DSS, GLBA) |

---

## 🏗 Architecture at a Glance

![Hybrid System Architecture — 3-Path Routing](docs/images/hybrid_architecture.png)

---

## 🔑 Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **LLM Engine** | Amazon Bedrock | Managed, multi-model (Claude 3.5, Llama 3), built-in guardrails, knowledge bases |
| **Voice** | Amazon Connect | Native Lex integration, Contact Lens analytics, live agent transfer |
| **Orchestrator** | Amazon Lex V2 (Assisted NLU) | LLM-enhanced classification, 3 intent types: Standard, QnAIntent, BedrockAgentIntent |
| **Routing** | Hybrid 3-path | 🟢 Deterministic (simple) · 🔵 Knowledge/FAQ (RAG) · 🟠 LLM Agent (complex) |
| **Safety** | Bedrock Guardrails | PII redaction, toxicity, hallucination grounding, denied topics |
| **IaC** | Terraform + CDK | Terraform for base infra, CDK for dynamic use case provisioning |
| **CI/CD** | GitLab CI/CD | Company standard; MR-based workflow with approval gates |
| **Vector Store** | OpenSearch Serverless | Serverless scaling, native Bedrock KB integration |
| **Observability** | CloudWatch + Athena + LLM Evaluator | 3-tier: real-time health, conversation analytics, AI-powered quality |

---

## 🚀 Quick Start for LOB Partners

1. Copy a use case template from `use-cases/` directory
2. Fill in intents, slots, knowledge sources, and guardrail preferences
3. Submit a merge request to GitLab
4. Platform auto-provisions your POC environment
5. Test, iterate, promote to production

See [Onboarding Pipeline](docs/02-onboarding-pipeline.md) for full details.

---

## 📂 Repository Structure

```
conversational-ai-platform/
├── README.md                  ← You are here
├── docs/                      # Architecture documentation
│   ├── 01-system-architecture.md
│   ├── 02-onboarding-pipeline.md
│   ├── 03-observability.md
│   ├── 04-security-compliance.md
│   └── images/
├── infra/                     # Terraform — base infrastructure
├── platform/                  # Platform CLI & CDK app
├── use-cases/                 # Use case specs (YAML)
├── lambdas/                   # Fulfillment Lambda functions
├── tests/                     # Test harnesses
└── .gitlab-ci.yml             # GitLab CI/CD pipeline
```
