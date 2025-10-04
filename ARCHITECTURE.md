# Procurement Management System Architecture

## Purpose & Scope

This document specifies the target architecture for the Procurement Management System (PMS) that supports enterprise construction procurement. It will evolve alongside the roadmap and acts as the contract between engineering, security, and the Procurement Center of Excellence.

## Architectural Principles

- **Modular services:** Decompose by business capability (intake, sourcing, contract, supplier, operations) to isolate change and scale selectively.

- **Event-driven cohesion:** Use a persistent event bus for data propagation, audit, and eventual consistency across domains.

- **Secure by default:** Enforce least-privilege access, encrypted transport/storage, and tamper-resistant audit trails for every transaction.

- **Cloud native reliability:** Automate scalability, fault tolerance, and observability via container orchestration, service mesh, and infrastructure as code.

- **Config over code:** Support regional policy variations through configurable workflows, rules, and templates to avoid forks.

- **Open-source first:** Prioritize open-source platforms and frameworks with active communities to preserve code ownership and avoid vendor lock-in; proprietary services require explicit exception approval.

## Technology Constraints

- Bi-directional integration with the corporate SAP landscape (S/4HANA or ECC) is mandatory for vendor master data, project/WBS alignment, commitments, goods movements, and financial postings.

- Core platform components should leverage open-source technologies (e.g., Kubernetes, Istio, PostgreSQL, Camunda, Keycloak, Kafka, Prometheus) with automation for updates and security hardening.

- Any proprietary or cloud-managed alternatives must document justification, cost, and exit strategy, and obtain Architecture Review Board sign-off.

## Logical View

| Layer | Components | Responsibilities |
| --- | --- | --- |
| Experience | Web client (internal roles), Supplier portal, Mobile receiving app | Task-focused UX, offline support for sites, localization |
| API Gateway | GraphQL facade, REST proxy, auth adapters | Protocol translation, request throttling, schema governance |
| Domain Services | Project Intake, Sourcing, Contract, Procurement Ops, Supplier, Inventory, Analytics, Notification | Encapsulate business logic, manage own data stores, expose APIs and events |
| Foundational Services | Workflow engine (Camunda), Rules engine (Drools), Identity service, Document service, Reporting service | Shared capabilities consumed across domains |
| Data | PostgreSQL clusters, Object storage (documents), Redis cache, Data lakehouse (Delta Lake) | Transactional persistence, search, analytical workloads |
| Integration | ERP/MRP adapters, BIM & scheduling adapters, HRIS connector, Logistics connectors | Inbound/outbound data synchronization |

## Physical Deployment

- **Kubernetes Cluster:** Multi-AZ managed cluster hosts stateless services; node pools tuned for compute- or IO-heavy workloads.

- **Service Mesh (Istio):** Handles mTLS, traffic policy, circuit breaking, and distributed tracing headers.

- **Data Stores:**
  - PostgreSQL in highly available configuration with leader/follower nodes per domain.
  - Object storage (S3-compatible) for contracts, supplier docs, RFx artifacts.
  - Redis cluster for caching master data, session state, and workflow tokens.
  - Delta Lake on cloud storage for analytical workloads; fed via CDC from operational DBs.

- **Event Streaming:** Apache Kafka (3-node minimum) with tiered storage and schema registry for topic governance.

- **CI/CD Pipeline:** Git-based trunk with branch protections; build runners publish containers to registry, apply Terraform for infra, and Helm charts for deployments.

## Runtime & Interaction Patterns

- Synchronous flows use REST/GraphQL calls behind the gateway; gateway enforces OAuth scopes and injects correlation IDs.
- Long-running processes orchestrated via BPM engine; tasks surface to users through worklist UI and trigger microservice callbacks.
- Domain events (e.g., `PO.Created`, `Invoice.Matched`, `Supplier.Onboarded`) published to Kafka topics; consumers update read models or trigger external integrations.
- File-heavy actions upload to document service which emits metadata events for downstream processing (e.g., legal review).
- Supplier portal interactions proxy through DMZ ingress with web application firewall rules and rate limiting.

## Data Architecture

- **Master Data:** Projects, suppliers, materials reside in dedicated schemas with Golden Record governance; synchronized with ERP via nightly reconciliation plus delta events.
- **Transaction Data:** Each microservice owns its relational schema; cross-service joins implemented via APIs or denormalized projections to analytics store.
- **Analytics:** ETL streams operational events into Delta Lake, enabling BI dashboards, ad hoc SQL, and ML workflows. KPI snapshots materialized back into Analytics service for low-latency dashboards.
- **Retention:** Procurement transactions retained minimum 10 years; document retention configurable per region; archival tier for closed projects.

## Security Architecture

- **Identity & Access:** Integrate corporate IdP via OIDC for interactive logins, OAuth2 client credentials for service-to-service; SCIM automates provisioning/deprovisioning.
- **Authorization:** Attribute-based rules combining role, project, spend authority, and geography. Policy engine centralizes evaluation and emits decision logs.
- **Encryption:** TLS 1.2+ for all transport; database encryption at rest; secrets managed via Vault with dynamic credentials and auto-rotation.
- **Audit & Compliance:** Append-only audit log service writes to immutable storage and pushes to SIEM; supports forensic queries and regulatory reporting.
- **Threat Protection:** WAF in front of public endpoints, anomaly detection via behavior analytics, automated lockout for repeated failed supplier logins.

## Resilience & Performance

- Target service SLOs: 99.9% availability for internal APIs, 99.5% for supplier portal.
- Horizontal Pod Autoscaler scales services based on CPU, memory, and custom business metrics (queue depth).
- Read replicas serve reporting queries; circuit breakers fallback to cached data when upstream unavailable.
- Disaster recovery replicates databases and Kafka topics across regions; runbooks define failover steps with RPO ≤ 5 minutes, RTO ≤ 30 minutes.
- Synthetic monitoring simulates critical journeys (RFQ publish, PO approve, GRN post) for early incident detection.

## Observability & Operations

- Logging standardized with JSON structure; aggregated via Fluent Bit to Elastic/OpenSearch.
- Metrics from Prometheus exported to Grafana dashboards; SLO alerts piped into PagerDuty and Teams channels.
- Distributed tracing (OpenTelemetry) instrumented across gateway and services.
- Runbooks maintained per service with auto-linked dashboards and alert thresholds.
- Feature flags managed through centralized service to enable progressive rollouts and A/B testing.

## Integration Strategy

- **SAP** (S/4HANA or ECC) integration via middleware exposing IDoc/REST; asynchronous updates for vendor master, project/WBS data, commitments, goods receipts, and payment status.
- **Project Scheduling** consume milestones to align procurement deadlines and auto-adjust safety stock.
- **Logistics/3PL** ingest shipment telemetry to update expected receipt times and trigger alerts.
- **E-signature** (DocuSign/Adobe Sign) orchestrated via Contract service with callback webhooks.
- **Compliance APIs** for sanctions, tax IDs, insurance validation invoked during onboarding.

## Non-Functional Requirements

- Accessibility: WCAG 2.1 AA compliance on web portals.
- Localization: Multi-language UI, currency conversion, tax rule configuration by country.
- Scalability: Support 5k concurrent users, 100 active sourcing events, and 10k PO lines/day with <2s median response.
- Data Privacy: GDPR compliant data subject access requests and deletion workflows.

## Evolution & Governance

- Architecture Review Board evaluates significant design changes; ADRs stored in repository.
- Quarterly technical debt review to reprioritize refactoring and infra improvements.
- Chaos engineering exercises semi-annually to validate failure modes and observability.
- Versioning strategy: Semantic versioning for APIs; backward compatibility maintained for at least two minor versions.

## Open Questions

- Should supplier scoring leverage existing corporate analytics platform or stay within PMS?
- Decision pending on adopting managed Kafka service vs. self-hosted cluster.
- Need assessment of edge connectivity for remote construction sites to finalize offline sync strategy.
