# buyer Agents

**Date:** 2025-11-12
**Status:** Planning / Ideation Phase

---

## Overview

This document outlines potential AI agent integrations for the buyer application. Agents can automate procurement workflows, provide intelligent decision support, and reduce manual data entry.

## Agent Ideas

### 1. Sourcing Agent

**Purpose:** Automate quote collection from vendors

**Capabilities:**
- **Email parsing:** Extract quote information from vendor emails
- **Web scraping:** Monitor vendor websites for price updates
- **API integration:** Automatically fetch quotes from vendor APIs
- **Quote creation:** Create structured quote entries in the database
- **Price change alerts:** Notify when vendor prices change significantly

**Workflow:**
1. Monitor vendor communication channels (email, API, web)
2. Extract product name, price, currency, and availability
3. Match to existing products in the database
4. Create quote entry with proper vendor/product associations
5. Flag anomalies or missing information for human review

**Technical Requirements:**
- Email integration (IMAP/Gmail API)
- HTML/PDF parsing for quote documents
- Natural language processing for unstructured text
- Product name matching/fuzzy search

---

### 2. Forex Agent

**Purpose:** Maintain up-to-date currency exchange rates

**Capabilities:**
- **Automatic rate updates:** Fetch latest rates from forex APIs
- **Historical tracking:** Maintain rate history for trend analysis
- **Rate alerts:** Notify on significant rate fluctuations
- **Multi-source aggregation:** Combine rates from multiple sources
- **Rate prediction:** Forecast future exchange rates

**Workflow:**
1. Schedule periodic API calls to forex data providers
2. Validate and store new rates in the database
3. Update converted prices on existing quotes
4. Send alerts when rates change beyond threshold
5. Generate forex trend reports

**Data Sources:**
- [exchangerate-api.com](https://www.exchangerate-api.com/)
- [currencyapi.com](https://currencyapi.com/)
- [fixer.io](https://fixer.io/)
- Central bank feeds (ECB, Fed, etc.)

---

### 3. Price Comparison Agent

**Purpose:** Intelligent quote analysis and recommendations

**Capabilities:**
- **Best price finder:** Identify lowest-cost vendor for each product
- **Total cost analysis:** Calculate shipping, taxes, discounts
- **Vendor reliability scoring:** Track vendor performance over time
- **Budget optimization:** Suggest product/vendor combinations within budget
- **Bulk discount detection:** Identify quantity-based savings

**Workflow:**
1. Analyze all active quotes for a product/specification
2. Consider non-price factors (lead time, reliability, warranty)
3. Calculate total cost of ownership
4. Generate ranked recommendations with justification
5. Highlight cost-saving opportunities

**ML Models:**
- Price trend prediction
- Vendor reliability classification
- Delivery time estimation

---

### 4. Requisition Assistant Agent

**Purpose:** Help users create well-formed requisitions

**Capabilities:**
- **Natural language input:** "I need 10 laptops under $1000 each"
- **Specification suggestion:** Recommend specs based on requirements
- **Automated quote gathering:** Trigger sourcing agent for items
- **Budget validation:** Ensure requisition stays within limits
- **Approval routing:** Suggest appropriate approvers based on amount/category

**Workflow:**
1. User provides natural language description of needs
2. Agent extracts quantities, specifications, budget constraints
3. Searches existing products or creates new specifications
4. Pre-populates requisition with suggested items
5. Runs initial quote comparison
6. Presents draft requisition for user approval

**Example Inputs:**
- "Need 5 laptops with 16GB RAM, budget $8000"
- "Get quotes for 100 USB-C cables from our approved vendors"
- "Create requisition for office furniture, desks and chairs for 10 people"

---

### 5. Vendor Intelligence Agent

**Purpose:** Maintain vendor profiles and relationships

**Capabilities:**
- **Contact discovery:** Find vendor contact information online
- **Performance tracking:** Monitor delivery times, quality issues
- **Risk assessment:** Flag vendors with payment/quality problems
- **Relationship management:** Track communication history
- **Alternative suggestions:** Recommend backup vendors

**Workflow:**
1. Aggregate data from quotes, deliveries, payments
2. Calculate vendor performance metrics
3. Scan news/reviews for vendor reputation changes
4. Update vendor risk scores
5. Alert buyers to vendor issues

**Metrics Tracked:**
- On-time delivery percentage
- Quote accuracy (quoted vs actual price)
- Response time to inquiries
- Quality issue frequency
- Payment terms offered

---

### 6. Compliance & Audit Agent

**Purpose:** Ensure procurement follows organizational policies

**Capabilities:**
- **Policy enforcement:** Check purchases against spending limits
- **Documentation verification:** Ensure quotes have required info
- **Duplicate detection:** Flag redundant requisitions
- **Audit trail generation:** Create detailed purchase histories
- **Spend analysis:** Identify maverick spending

**Workflow:**
1. Monitor all requisitions and quotes
2. Apply business rules (spending limits, approved vendors, etc.)
3. Flag policy violations before approval
4. Generate compliance reports
5. Suggest corrective actions

**Policy Examples:**
- "Purchases >$5000 require 3 quotes"
- "Must use approved vendor list for IT equipment"
- "International vendors require compliance check"

---

### 7. Market Intelligence Agent

**Purpose:** Track market trends and pricing benchmarks

**Capabilities:**
- **Price benchmarking:** Compare internal quotes to market averages
- **Trend detection:** Identify inflation/deflation in categories
- **Demand forecasting:** Predict future procurement needs
- **Supply chain monitoring:** Track shortages and disruptions
- **Contract optimization:** Suggest bulk purchase opportunities

**Workflow:**
1. Aggregate external market data
2. Compare internal purchase prices to benchmarks
3. Identify outliers (paying too much or suspiciously low)
4. Generate market reports and alerts
5. Recommend strategic sourcing opportunities

**Data Sources:**
- Industry price indices
- Commodity market data
- Competitor intelligence
- Supply chain news feeds

---

### 8. Document Processing Agent

**Purpose:** Extract structured data from procurement documents

**Capabilities:**
- **Invoice parsing:** Extract line items from vendor invoices
- **Quote PDF analysis:** Convert PDF quotes to structured data
- **Specification extraction:** Parse technical spec sheets
- **Contract analysis:** Identify key terms and obligations
- **OCR for scanned docs:** Handle paper documents

**Workflow:**
1. Receive document (email attachment, upload, scan)
2. Classify document type (quote, invoice, spec sheet)
3. Extract structured fields (vendor, items, prices, dates)
4. Validate extracted data
5. Create/update database entries
6. Flag ambiguous data for human review

**Technologies:**
- OCR (Tesseract, Google Vision API)
- PDF parsing (pdfplumber, PyPDF2)
- NLP for text extraction
- Layout analysis for tables

---

### 9. Negotiation Assistant Agent

**Purpose:** Support buyers during vendor negotiations

**Capabilities:**
- **Historical price analysis:** Show past prices paid
- **Leverage identification:** Highlight buyer's negotiating power
- **Alternative options:** Present competing vendor quotes
- **Talking points generation:** Suggest negotiation strategies
- **Contract term suggestions:** Recommend favorable terms

**Workflow:**
1. Buyer initiates negotiation with vendor
2. Agent provides price history and market data
3. Suggests target price based on analysis
4. Tracks negotiation progress
5. Recommends when to accept or counter

**Insights Provided:**
- "You paid 15% less last year"
- "Competitor offers same product for 20% less"
- "High volume - request bulk discount"
- "Market prices trending down - wait or negotiate"

---

### 10. Chatbot Interface Agent

**Purpose:** Natural language interface for buyer application

**Capabilities:**
- **Query data:** "What did we pay for laptops last month?"
- **Create entries:** "Add a new vendor called TechCorp"
- **Run reports:** "Show me all expired quotes"
- **Guided workflows:** Walk users through complex processes
- **Contextual help:** Provide usage instructions

**Example Interactions:**
```
User: "Show me quotes for Apple laptops"
Agent: "Found 5 quotes for Apple products. Lowest: $1,299 from B&H Photo"

User: "Create a quote from that vendor for $1,299"
Agent: "I'll need the product name. Is it 'MacBook Pro 14-inch'?"

User: "Yes"
Agent: "Quote created. ID: 42. Want to add it to a requisition?"
```





---

## Agent Frameworks & SDKs

### 1. Google Agent Development Kit (ADK)

**Language:** Go, Python, TypeScript

Agent Development Kit (ADK) is a flexible and modular framework for developing and deploying AI agents. While optimized for Gemini and the Google ecosystem, ADK is model-agnostic, deployment-agnostic, and is built for compatibility with other frameworks. ADK was designed to make agent development feel more like software development, to make it easier for developers to create, deploy, and orchestrate agentic architectures that range from simple tasks to complex workflows.

**Links:**
- [ADK Documentation](https://google.github.io/adk-docs/)
- [ADK for Go](https://google.github.io/adk-docs/get-started/go/)

**Pros:**
- Native Go support (matches buyer's tech stack)
- Model-agnostic (not locked to Gemini)
- Strong Google ecosystem integration
- Built for production deployment

**Cons:**
- Newer framework (less mature ecosystem)
- Documentation still evolving
- Smaller community compared to LangChain

**Best For:** Sourcing Agent, Forex Agent, Document Processing Agent

---

### 2. LangChain / LangGraph

**Language:** Python (primary), Go (langchaingo)

Industry-standard framework for building LLM applications. LangGraph extends LangChain with stateful, multi-actor workflows using graph-based execution.

**Links:**
- [LangChain](https://python.langchain.com/)
- [LangGraph](https://langchain-ai.github.io/langgraph/)
- [LangChain Go](https://github.com/tmc/langchaingo)

**Pros:**
- Massive ecosystem of integrations
- Excellent documentation and community
- Built-in tools for RAG, agents, chains
- Strong production deployment story (LangSmith)

**Cons:**
- Python-first (Go port less feature-complete)
- Can be complex for simple use cases
- Performance overhead from abstraction layers

**Best For:** Chatbot Interface Agent, Requisition Assistant Agent, Market Intelligence Agent

---

### 3. Anthropic Claude SDK

**Language:** Go, Python, TypeScript

Official SDK for Claude API with support for function calling, streaming, and vision.

**Links:**
- [Anthropic Go SDK](https://github.com/anthropics/anthropic-sdk-go)
- [Anthropic API Docs](https://docs.anthropic.com/)

**Pros:**
- Direct API access (minimal abstraction)
- Excellent Go support
- High-quality function calling
- Strong reasoning capabilities

**Cons:**
- Lower-level than frameworks
- Need to build agent orchestration yourself
- Claude-specific (vendor lock-in)

**Best For:** Price Comparison Agent, Negotiation Assistant Agent, Compliance Agent

---

### 4. OpenAI SDK with Assistants API

**Language:** Go, Python, JavaScript

OpenAI's official SDKs with support for Assistants API (stateful agents with built-in tools).

**Links:**
- [OpenAI Go SDK](https://github.com/sashabaranov/go-openai)
- [Assistants API](https://platform.openai.com/docs/assistants/overview)

**Pros:**
- Managed state and memory
- Built-in code interpreter, file search, function calling
- Easy to get started
- Well-documented

**Cons:**
- OpenAI vendor lock-in
- Assistants API can be expensive
- Less control over agent behavior

**Best For:** Chatbot Interface Agent, Document Processing Agent

---

### 5. CrewAI

**Language:** Python

Multi-agent framework for coordinating multiple specialized agents as a "crew" working together.

**Links:**
- [CrewAI](https://github.com/joaomdmoura/crewAI)
- [CrewAI Docs](https://docs.crewai.com/)

**Pros:**
- Built for multi-agent orchestration
- Role-based agent design
- Task delegation and collaboration
- Simple, intuitive API

**Cons:**
- Python only (no Go support)
- Newer project (less battle-tested)
- Would require separate service

**Best For:** Complex workflows requiring multiple agents (Sourcing + Price Comparison + Vendor Intelligence)

---

### 6. Semantic Kernel (Microsoft)

**Language:** Go, Python, C#

Microsoft's SDK for AI orchestration, emphasizing "semantic functions" and memory.

**Links:**
- [Semantic Kernel](https://github.com/microsoft/semantic-kernel)
- [Semantic Kernel Go](https://github.com/microsoft/semantic-kernel-go)

**Pros:**
- Good Go support
- Azure integration
- Enterprise-focused
- Plugin system for extensibility

**Cons:**
- Azure-centric design
- Less community momentum than LangChain
- Heavier abstractions

**Best For:** Enterprise deployments with Azure infrastructure

---

### 7. AutoGPT / AgentGPT

**Language:** Python, TypeScript

Autonomous agents that can break down goals into tasks and execute them.

**Links:**
- [AutoGPT](https://github.com/Significant-Gravitas/AutoGPT)
- [AgentGPT](https://github.com/reworkd/AgentGPT)

**Pros:**
- Fully autonomous operation
- Goal-oriented design
- Web UI included (AgentGPT)

**Cons:**
- Can be unpredictable
- High token usage
- Requires careful guardrails

**Best For:** Experimental / R&D projects

---

### 8. Haystack (deepset)

**Language:** Python

Open-source framework for building search and NLP applications with LLMs.

**Links:**
- [Haystack](https://haystack.deepset.ai/)

**Pros:**
- Excellent for RAG pipelines
- Document processing focus
- Production-ready
- Strong search capabilities

**Cons:**
- Python only
- More focused on search/RAG than general agents

**Best For:** Document Processing Agent, Market Intelligence Agent (with RAG over market data)

---

### 9. Custom Go Implementation

**Language:** Go

Build agents directly using Go with LLM API clients and custom orchestration logic.

**Approach:**
- Use `anthropic-sdk-go` or `go-openai` for LLM calls
- Implement tool/function calling manually
- Use `gorm.DB` directly for data access
- Build simple state machine for agent workflows

**Pros:**
- Full control over behavior
- Native integration with buyer codebase
- No external dependencies
- Optimal performance

**Cons:**
- More development effort
- Need to implement orchestration, memory, etc.
- Maintenance burden

**Best For:** Simple, high-performance agents (Forex Agent, Price Comparison Agent)

---

### 10. LangChain + Go Microservices Hybrid

**Language:** Python (agents) + Go (API/services)

Run agent logic in Python (LangChain) as a separate service, exposed via HTTP API to the Go application.

**Architecture:**
```
buyer (Go) <--> HTTP API <--> Agent Service (Python/LangChain)
```

**Pros:**
- Leverage LangChain's ecosystem
- Keep buyer in Go
- Clear separation of concerns
- Can use different models per agent

**Cons:**
- Additional deployment complexity
- Network latency
- Two codebases to maintain

**Best For:** Complex agents needing LangChain's rich ecosystem

---

### 11. Eino (ByteDance CloudWeGo)

**Language:** Go

Eino is an AI application development framework from ByteDance's CloudWeGo project. It provides a comprehensive toolkit for building LLM-powered applications with native Go support, focusing on production readiness and enterprise-grade features.

**Links:**
- [Eino GitHub](https://github.com/cloudwego/eino)
- [CloudWeGo](https://www.cloudwego.io/)

**Pros:**
- Pure Go implementation (no CGO)
- Built by ByteDance for production scale
- Strong focus on performance and reliability
- Native support for multiple LLM providers
- Modular architecture (compose, flow, retrieval)
- Active development and enterprise backing
- Integration with ByteDance's ecosystem

**Cons:**
- Newer project (less mature documentation)
- Smaller English-language community (Chinese-first)
- Less third-party integrations compared to LangChain
- Still evolving API

**Best For:** Production Go applications needing enterprise-grade reliability (All agents in Phase 1-2)

**Technical Highlights:**
- **Compose:** Chain LLM calls with structured workflows
- **Flow:** Build complex multi-step agent pipelines
- **Retrieval:** Built-in RAG support with vector stores
- **Callbacks:** Observability and monitoring hooks
- **Streaming:** Native support for streaming responses

**Example Use Case:**
```go
// Eino-powered Forex Agent
import "github.com/cloudwego/eino/compose"

agent := compose.NewChain(
    retriever,  // Get latest rates
    llm,        // Validate and process
    callback,   // Log to buyer
)
```

---

### 12. Lindy (Lindy.ai Framework)

**Language:** Go, TypeScript

Lindy is an open-source framework for building autonomous AI agents with a focus on real-world business automation. It emphasizes long-running tasks, external tool integration, and reliability.

**Links:**
- [Lindy.ai](https://www.lindy.ai/)
- [Lindy Framework (GitHub)](https://github.com/lindyai/lindy) *(hypothetical - check for actual repo)*

**Pros:**
- Designed for business automation
- Long-running task support
- External API integration focus
- Scheduling and retry logic built-in
- Email/calendar/webhook integrations
- Human approval workflows

**Cons:**
- Primarily commercial product (limited open-source)
- Less flexible than general frameworks
- More opinionated architecture
- Potential vendor lock-in

**Best For:** Business automation agents (Sourcing Agent, Vendor Intelligence Agent)

**Key Features:**
- **Triggers:** Schedule, webhook, email-based activation
- **Actions:** 200+ pre-built integrations
- **Conditions:** Business logic and routing
- **Approvals:** Human-in-the-loop workflows

---

### 13. Rivet (Ironclad/Rivet)

**Language:** TypeScript (with Go bindings possible)

Rivet is a visual programming environment for creating AI agents and workflows. It provides a node-based editor for designing complex agent behavior without code.

**Links:**
- [Rivet GitHub](https://github.com/Ironclad/rivet)
- [Rivet Documentation](https://rivet.ironcladapp.com/)

**Pros:**
- Visual workflow designer
- Great for prototyping
- Export to code
- Built-in debugging tools
- Version control for workflows
- Team collaboration features

**Cons:**
- TypeScript/JavaScript focused
- Limited Go support
- May require Node.js runtime
- Less suitable for complex logic

**Best For:** Rapid prototyping and non-technical stakeholder collaboration

---

### 14. Superagent

**Language:** Python, TypeScript

Open-source framework for building, deploying, and managing LLM agents with a focus on production deployments and API-first design.

**Links:**
- [Superagent GitHub](https://github.com/homanp/superagent)
- [Superagent Cloud](https://www.superagent.sh/)

**Pros:**
- Built for production deployment
- API-first architecture (works with Go via HTTP)
- Agent marketplace and templates
- Built-in vector database
- Memory management
- Usage tracking and analytics

**Cons:**
- Python/TypeScript (no native Go)
- Would require separate service
- Commercial features behind paywall
- Smaller community than LangChain

**Best For:** Rapid deployment of production agents as microservices

---

### 15. SwarmGo

**Language:** Go

SwarmGo is a Go implementation inspired by OpenAI's Swarm framework for building multi-agent systems. It provides lightweight, ergonomic orchestration of multiple agents with handoffs and context management.

**Links:**
- [SwarmGo GitHub](https://github.com/prathyushnallamothu/swarmgo)
- [OpenAI Swarm (Python original)](https://github.com/openai/swarm)

**Pros:**
- Pure Go implementation (native to buyer stack)
- Lightweight and minimal abstractions
- Multi-agent coordination with handoffs
- Context management across agents
- Educational and easy to understand
- No external dependencies beyond LLM APIs
- Good for experimenting with agent patterns

**Cons:**
- Newer project (less battle-tested)
- Smaller community and ecosystem
- Limited documentation compared to mature frameworks
- Experimental/educational focus (not production-hardened)
- Less feature-rich than LangChain or ADK

**Best For:** Learning multi-agent patterns, prototyping agent interactions in pure Go

**Technical Highlights:**
- **Agent Handoffs:** Agents can transfer control to other specialized agents
- **Context Variables:** Shared state across agent interactions
- **Function Calling:** Direct integration with LLM tool/function calling
- **Minimal API:** Simple, understandable interface

**Example Use Case:**
```go
// SwarmGo multi-agent procurement workflow
import "github.com/prathyushnallamothu/swarmgo"

// Define specialized agents
sourcingAgent := swarmgo.Agent{
    Name: "Sourcing Agent",
    Instructions: "Find and collect vendor quotes",
    Functions: []swarmgo.Function{collectQuotes, searchVendors},
}

priceAgent := swarmgo.Agent{
    Name: "Price Comparison Agent",
    Instructions: "Analyze quotes and recommend best option",
    Functions: []swarmgo.Function{compareQuotes, calculateTotal},
}

// Agent handoff workflow
response := swarmgo.Run(
    client,
    sourcingAgent,
    "Find quotes for 10 laptops under $1000",
    contextVariables,
)
```

**Recommendation for buyer:**
- **Good for:** Prototyping multi-agent patterns in Phase 1-2
- **Use case:** Experiment with Sourcing Agent → Price Comparison Agent workflows
- **Caution:** Consider more mature frameworks (Eino, ADK) for production
- **Best approach:** Use SwarmGo for learning, then migrate to Eino or ADK for production deployment

---

## Tentative Roadmap

### Phase 1: Foundation (Months 1-2)

**Goal:** Prove value with simple, high-ROI agents

**Agents to Implement:**
1. **Forex Agent** (Week 1-2)
   - Simplest agent to build
   - Clear, measurable value
   - Low risk of errors

2. **Price Comparison Agent** (Week 3-4)
   - Leverages existing data
   - Provides immediate decision support
   - No external integrations needed

**Technology Stack:**
- **Framework:** Eino (CloudWeGo) OR Custom Go + Anthropic Claude SDK
- **Recommended:** **Eino** for production-grade structure
- **Rationale:**
  - Stay in Go (no Python dependency)
  - Eino provides enterprise-proven framework with modularity
  - Direct API calls (simple, fast)
  - Full control over behavior
  - Minimal external dependencies
  - Eino scales better if adding more agents

**Infrastructure:**
- Add `internal/agents/` package
- Create agent interface: `Agent`, `Run()`, `GetTools()`
- Implement simple in-memory state management
- Add agent metrics/logging

**Deliverables:**
- Working Forex Agent with scheduler
- Price Comparison Agent accessible via CLI/web
- Agent observability (logs, metrics)
- Documentation

---

### Phase 2: Automation (Months 3-4)

**Goal:** Reduce manual data entry

**Agents to Implement:**
3. **Document Processing Agent** (Week 5-8)
   - Parse PDF/email quotes
   - Extract structured data
   - Auto-create quote entries

4. **Sourcing Agent** (Week 9-12)
   - Monitor vendor emails
   - Web scraping for price updates
   - API integrations

**Technology Stack:**
- **Framework:** Google ADK (Go)
- **Rationale:**
  - Native Go support
  - Better for complex workflows
  - Model-agnostic (can switch LLMs)
  - Production-ready orchestration

**Infrastructure:**
- Email integration (IMAP/Gmail API)
- Document storage (S3 or local filesystem)
- Queue system for async processing (Go channels or Redis)
- Human-in-the-loop review interface

**Deliverables:**
- Upload quote PDFs via web interface
- Email monitoring for quotes
- Review queue for extracted quotes
- Agent accuracy metrics

---

### Phase 3: Intelligence (Months 5-6)

**Goal:** Add decision support and insights

**Agents to Implement:**
5. **Vendor Intelligence Agent** (Week 13-16)
   - Track vendor performance
   - Risk assessment
   - Alternative suggestions

6. **Market Intelligence Agent** (Week 17-20)
   - Price benchmarking
   - Trend detection
   - Market alerts

**Technology Stack:**
- **Framework:** LangChain (Python microservice) + Go API
- **Rationale:**
  - Need RAG for market data
  - LangChain's ecosystem for external data
  - Separate service isolates complexity

**Infrastructure:**
- Vector database (Chroma or Qdrant) for RAG
- Python agent service with HTTP API
- Data pipelines for external market data
- Analytics dashboard

**Deliverables:**
- Vendor performance dashboards
- Market intelligence reports
- Price alerts and recommendations
- API documentation for agent service

---

### Phase 4: Conversational Interface (Months 7-8)

**Goal:** Natural language interface for users

**Agents to Implement:**
7. **Chatbot Interface Agent** (Week 21-24)
   - Natural language queries
   - Data entry via chat
   - Guided workflows

8. **Requisition Assistant Agent** (Week 25-28)
   - NL requisition creation
   - Spec suggestions
   - Budget validation

**Technology Stack:**
- **Framework:** OpenAI Assistants API
- **Rationale:**
  - Managed state and memory
  - Built-in function calling
  - Easy integration

**Infrastructure:**
- Chat interface in web UI
- Conversation history storage
- Function calling to buyer services
- User feedback system

**Deliverables:**
- Chat widget in web interface
- Voice-enabled requisition creation
- Conversation history
- User satisfaction metrics

---

### Phase 5: Advanced Capabilities (Months 9-12)

**Goal:** Complex multi-agent workflows

**Agents to Implement:**
9. **Negotiation Assistant Agent**
10. **Compliance & Audit Agent**

**Technology Stack:**
- **Framework:** CrewAI or LangGraph (Python service)
- **Rationale:**
  - Multi-agent coordination needed
  - Complex decision workflows
  - Role-based collaboration

**Infrastructure:**
- Multi-agent orchestration
- Long-term memory/knowledge base
- Integration with approval workflows
- Audit trail system

**Deliverables:**
- Negotiation support tools
- Automated compliance checking
- Audit report generation
- Multi-agent workflow visualization

---

## Technical Recommendations

### Recommended Architecture

**Hybrid Approach:**
```
┌─────────────────────────────────────────────────────────┐
│ buyer Application (Go)                                  │
│                                                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │ CLI / Web / GUI Interfaces                       │   │
│  └──────────────┬───────────────────────────────────┘   │
│                 │                                       │
│  ┌──────────────┴───────────────────────────────────┐   │
│  │ Agent Coordinator (Go)                           │   │
│  │ - Agent registry                                 │   │
│  │ - Routing layer                                  │   │
│  │ - Observability                                  │   │
│  └──────┬──────────────────────────────┬────────────┘   │
│         │                              │                │
│  ┌──────┴─────────┐            ┌───────┴──────────┐     │
│  │ Simple Agents  │            │ Complex Agents   │     │
│  │ (Native Go)    │            │ (Microservice)   │     │
│  │                │            │                  │     │
│  │ - Forex        │            │ - Document Proc  │     │
│  │ - Price Comp   │            │ - Sourcing       │     │
│  └────────────────┘            │ - Market Intel   │     │
│                                │ - Chatbot        │     │
│                                └──────────────────┘     │
└─────────────────────────────────────────────────────────┘
```

### Technology Choices by Phase

| Phase | Primary Tech | Alternative | Rationale |
|-------|-------------|------------|-----------|
| Phase 1 | Custom Go + Claude SDK | **Eino (Go)** | Simple, fast, native. Eino offers better structure if scaling |
| Phase 2 | Google ADK (Go) | **Eino (Go)** | Document processing, workflows. Eino for production-grade reliability |
| Phase 3 | LangChain (Python) | **Eino + Vector DB (Go)** | RAG, external data. Eino if staying pure Go |
| Phase 4 | OpenAI Assistants | Eino + Custom Memory | Managed state, easy chat |
| Phase 5 | CrewAI/LangGraph | - | Multi-agent orchestration |

**Note:** **Eino (ByteDance CloudWeGo)** is a strong alternative for Phases 1-3 if you want to stay in pure Go. It provides enterprise-grade features with better structure than custom implementations, while avoiding Python dependencies.

### Framework Comparison for buyer

**Recommended Stack Evolution:**

```
Phase 1-2: Eino (Go) OR Custom Go + Claude SDK
├─ Pure Go implementation
├─ Production-ready from ByteDance
├─ Good for Forex, Price Comparison, Document Processing
└─ Scales to complex workflows

Phase 3: Eino (Go) OR LangChain (Python)
├─ If pure Go: Eino + vector DB
├─ If need rich ecosystem: LangChain service
└─ Decision point based on Phase 1-2 learnings

Phase 4-5: OpenAI Assistants OR CrewAI
├─ Conversational: OpenAI Assistants (simplest)
├─ Multi-agent: CrewAI/LangGraph
└─ Could use Eino if building custom orchestration
```

**Why Eino is Compelling for buyer:**
1. **Same Language:** Stay in Go, leverage existing buyer codebase
2. **Enterprise-Proven:** Battle-tested at ByteDance scale
3. **Performance:** No Python interpreter overhead
4. **Simple Deployment:** Single binary with buyer
5. **Modular:** Compose, Flow, Retrieval can be adopted incrementally
6. **Production-Ready:** Built-in observability, retry logic, streaming

**When to Choose Python Frameworks:**
- Need LangChain's massive ecosystem (1000+ integrations)
- Complex RAG pipelines with specialized retrievers
- Multi-agent coordination (CrewAI strength)
- Rapid prototyping with established tools

### Development Principles

1. **Start Simple**
   - Implement Forex Agent first (lowest complexity)
   - Prove value before scaling
   - Avoid over-engineering

2. **Stay in Go When Possible**
   - Leverage existing codebase
   - Better performance
   - Simpler deployment
   - Only use Python for complex needs (RAG, multi-agent)

3. **Human-in-the-Loop**
   - All agent actions require human approval initially
   - Build confidence before full automation
   - Always provide explanation for agent decisions

4. **Observability First**
   - Log all agent actions
   - Track costs (token usage)
   - Monitor accuracy
   - Measure user satisfaction

5. **Incremental Deployment**
   - Feature flags for agent features
   - A/B testing for agent vs manual workflows
   - Gradual rollout to users

### Infrastructure Requirements

**Phase 1-2 (Simple Agents):**
- Go 1.21+
- Existing buyer infrastructure
- LLM API keys (Anthropic Claude)
- Minimal additional dependencies

**Phase 3-4 (Python Agents):**
- Python 3.11+ service
- Vector database (Chroma/Qdrant)
- Redis for queuing
- Additional 2-4GB RAM

**Phase 5 (Multi-Agent):**
- Kubernetes for orchestration (optional)
- Monitoring (Prometheus, Grafana)
- Separate database for agent state
- ~$100-500/month LLM API costs

### Cost Estimates

**Development Time:**
- Phase 1: 2 months (1 developer)
- Phase 2: 2 months (1 developer)
- Phase 3: 2 months (1-2 developers)
- Phase 4: 2 months (1-2 developers)
- Phase 5: 4 months (2 developers)
- **Total:** 12 months, ~1.5 FTE average

**LLM API Costs (estimated monthly):**
- Phase 1: $10-50 (periodic tasks)
- Phase 2: $50-200 (document processing)
- Phase 3: $100-500 (RAG queries)
- Phase 4: $200-1000 (chat interactions)
- Phase 5: $300-1500 (multi-agent workflows)

### Success Metrics

**Phase 1:**
- Forex rates updated daily without manual intervention
- Price comparison agent used in 50%+ of quote evaluations

**Phase 2:**
- 70%+ quote PDFs automatically extracted
- 50% reduction in manual quote entry time

**Phase 3:**
- Vendor risk scores calculated for 100% of vendors
- 3+ market insights generated per week

**Phase 4:**
- 30%+ of requisitions created via chatbot
- 80%+ user satisfaction with NL interface

**Phase 5:**
- Automated compliance checks on 100% of requisitions
- Negotiation agent used in 50%+ of high-value purchases

### Risk Mitigation

1. **LLM Accuracy**
   - Always validate agent outputs
   - Human review for high-value decisions
   - Gradual confidence thresholds

2. **Cost Control**
   - Set monthly budget limits
   - Monitor token usage
   - Cache common queries
   - Use cheaper models for simple tasks

3. **Vendor Lock-in**
   - Use model-agnostic frameworks (ADK, LangChain)
   - Abstract LLM calls behind interface
   - Test with multiple providers

4. **Security**
   - Sanitize agent inputs/outputs
   - Audit agent actions
   - Rate limiting
   - No PII in prompts without encryption

### Alternatives Considered

**Go-Only Approach:**
- **Pros:** Single language, simpler deployment
- **Cons:** Limited ecosystem, harder to implement complex workflows
- **Decision:** Use Go for simple agents, Python for complex ones

**Full Python Rewrite:**
- **Pros:** Rich agent ecosystem
- **Cons:** Massive refactor, performance hit
- **Decision:** Keep buyer in Go, add Python agents as services

**Commercial Agent Platforms:**
- **Examples:** Relevance AI, Dust, Fixie
- **Pros:** Managed infrastructure, faster time-to-market
- **Cons:** Vendor lock-in, cost, less control
- **Decision:** Build custom for flexibility and learning

---

## Next Steps

1. **Validate Assumptions**
   - Survey users: Which agents would provide most value?
   - Analyze current workflows: Where is manual work most painful?
   - Prototype Forex Agent: Prove technical feasibility

2. **Set Up Infrastructure**
   - Add `internal/agents/` package structure
   - Set up LLM API accounts (Anthropic, OpenAI)
   - Configure observability (logs, metrics)

3. **Start Phase 1**
   - Implement Forex Agent
   - Measure success
   - Get user feedback
   - Iterate before Phase 2

4. **Document Learnings**
   - What worked / what didn't
   - Cost actuals vs estimates
   - Adjust roadmap based on results

---

## Conclusion

Agent integration can significantly enhance buyer's capabilities, but should be approached incrementally. Start with simple, high-value agents (Forex, Price Comparison) using native Go implementations. As confidence grows, add more complex agents with appropriate frameworks (ADK, LangChain, OpenAI).

The hybrid architecture (Go + Python microservices) balances the need for a rich agent ecosystem with the performance and maintainability of the existing Go codebase.

Key success factors:
- Start simple and prove value
- Human-in-the-loop for safety
- Strong observability
- Incremental deployment
- Cost monitoring

With this approach, buyer can evolve from a data management tool to an intelligent procurement assistant over 12 months.