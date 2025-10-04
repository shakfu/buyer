# Procurement Management System Design

## Executive Summary

The Procurement Management System (PMS) centralizes sourcing, contracting, and material logistics for large construction firms. It equips the Head Procurement Manager (HPM) with real-time visibility into spend, supplier risk, and project readiness while streamlining daily work for category leads, site buyers, and supplier partners. Optimized procurement through guided workflows, analytics, and integrated compliance controls reduces delays, lowers total cost of ownership, and improves delivery confidence for capital projects.

## Strategic Objectives

- Cut cycle time from requisition to award by standardizing intake and bidding.
- Reduce material and services spend via data-driven supplier selection and negotiation support.
- Strengthen compliance with contract, safety, and regulatory policies through automated approvals.
- Provide portfolio-level KPIs for the HPM, exposing risks before they impact project schedules.
- Maintain auditability and traceability of every procurement decision.

## Personas & Needs

- **Head Procurement Manager:** Portfolio dashboards, exception alerts, approval queues, policy governance.
- **Category Leads:** Sourcing workbench, supplier scorecards, contract authoring with clause library.
- **Site Buyers & Storekeepers:** Mobile-friendly receiving, variance logging, rapid PO call-offs, inventory views.
- **Project Managers:** Visibility into procurement milestones and risks aligned to project schedules.
- **Finance & AP:** 3-way match status, accrual snapshots, dispute resolution workspace.
- **Suppliers:** Secure portal for RFx participation, PO acknowledgments, ASN updates, compliance docs.

## System Scope

- Demand intake with project coding and approval gates.
- Sourcing events (RFI, RFQ, e-auction), bid evaluation, and award decisioning.
- Contract lifecycle management with template library and version control.
- Supplier onboarding, qualification, and performance management.
- Purchase order and release execution with delivery milestone tracking.
- Site receiving, inspection, and non-conformance management.
- Invoice capture, 3-way matching, and discrepancy handling.
- Spend analytics, forecasting, and risk reporting.

## Technology & Integration Constraints

- The PMS must interface bi-directionally with the company SAP estate (S/4HANA or ECC), covering vendor master, project/WBS data, commitments, goods movements, and payment status.
- Implementation favors open-source components to preserve IP ownership; proprietary services require documented exceptions approved by the Architecture Review Board.
- Self-hosted deployments should default to community or enterprise-supported OSS distributions (e.g., Kubernetes, PostgreSQL, Camunda, Keycloak) with automation for patching and security updates.

## Core Workflows

1. **Project Launch Intake:** Project manager submits procurement needs; routing engine applies approval matrix; upon approval, sourcing tasks auto-generated.
2. **Sourcing Event:** Category lead drafts RFx, invites qualified suppliers, collects bids, applies weighted scoring, and submits award for HPM approval.
3. **Contract Authoring:** Select template, merge negotiated clauses, route for legal review, execute via e-signature, and store in contract repository.
4. **Purchase Execution:** Raise PO or release frame agreement, schedule deliveries, notify logistics partners, track acknowledgments.
5. **Receiving & Variance:** Site buyer records receipts (web or mobile), flags shortages/damages, triggers corrective workflows.
6. **Invoice & Match:** Capture invoice (EDI or manual), run automated 3-way match, escalate disputes to finance/shared services.
7. **Supplier Review:** Periodic quality/OTD scoring, issue corrective actions, update risk profile feeding sourcing eligibility.

## Domain Model Overview

- **Project** ↔ **Demand** (1:M) tying construction phases to procurement lines.
- **Supplier** ↔ **Qualification**, **Contract**, **RFx Response**, **PerformanceMetric**.
- **RFx Event** with **LineItem**, **Bid**, **Scorecard**, **Attachment**.
- **Contract** producing **PO** and **Release** records with **DeliveryMilestone**.
- **InventoryLocation** tracking **Receipt**, **InspectionReport**, **NonConformance**.
- **Invoice** linked to **PO**, **Receipt**, **PaymentStatus**.
- **User**, **Role**, and **ApprovalPolicy** determine task routing; all actions recorded in **AuditLog**.

## Architecture Overview
- **Services:** Project Intake, Sourcing, Contract, Procurement Operations, Supplier, Inventory, Finance Integration, Analytics, Notification.
- **API Gateway:** Unified REST/GraphQL surface with OAuth2 scopes; throttling and schema mediation.
- **Event Bus:** Kafka topics for PO events, delivery updates, invoice status, audit streams enabling integrations and async workflows.
- **Workflow & Rules:** BPM engine (e.g., Camunda) orchestrates long-running processes; rules engine (Drools) enforces approval thresholds and policy checks.
- **Datastores:** Operational data in PostgreSQL clusters (partitioned by tenant); documents in S3-compatible store; analytical data lakehouse (Delta Lake) fed via CDC.
- **Client Experience:** Responsive web app for internal roles, hardened supplier portal, optional mobile app for site receiving leveraging offline sync.
- **Infrastructure:** Containerized services on Kubernetes with IaC (Terraform) and service mesh (Istio) for observability and zero-trust networking.

## Integrations

- SAP (S/4HANA or ECC) for GL postings, vendor master sync, project/WBS references, goods movement updates, and payment status.
- Project management (Primavera/MS Project/BIM tools) for schedule dependencies.
- HRIS for org hierarchy and delegation-of-authority data.
- Logistics and 3PL telemetry for delivery ETA updates.
- DMS and e-signature (SharePoint, DocuSign) for contract artifacts.
- Compliance services (sanctions lists, tax validation) and corporate SIEM for security monitoring.

## Security & Compliance

- Enterprise SSO (SAML/OIDC) with SCIM lifecycle management; MFA enforced for high-risk roles.
- Role-based access with segregation-of-duties policies (e.g., same user cannot approve and receive).
- Field-level encryption for supplier banking/PII; KMS-managed keys.
- Immutable audit logging streamed to both PMS and corporate SIEM.
- Configurable approval thresholds, dual approval for high-value or emergency procurements.
- Data residency controls and retention policies aligned with regional regulations.

## Performance & Reliability

- Horizontal auto-scaling for ingest-heavy services (RFx, events) with HPA on CPU and queue depth.
- Multi-AZ PostgreSQL with async replicas; read replicas serve analytics queries.
- Event bus mirrors across regions; RPO ≤ 5 minutes, RTO ≤ 30 minutes.
- Application caching (Redis) for master data and KPI tiles; back-pressure mechanisms on supplier portal to prevent cascading failures.

## Analytics & Optimization

- KPI suite: savings vs. target, supplier OTIF, cycle times, contract leakage, inventory turns.
- ML models for supplier risk scoring, demand forecasting, and anomaly detection in invoices.
- Optimization heuristics (linear programming) suggest award splits based on cost, risk, and logistics constraints.
- Predictive alerts surface to HPM dashboard with recommended actions.

## Operations & Governance

- CI/CD with automated unit, contract, integration, security scans, and performance tests.
- Observability stack: Prometheus/Grafana, distributed tracing, centralized logging with alert runbooks.
- Change Advisory Board oversight for major releases; blue-green or canary deployments for high-risk features.
- Procurement Center of Excellence defines master data governance, clause library updates, and policy revisions.

## Product Roadmap

1. **Phase 0 – Discovery:** Shadow procurement processes, align data sources, define MVP metrics.
2. **Phase 1 – Core Execution:** Deliver intake, sourcing, PO, receiving, and basic dashboards for pilot projects.
3. **Phase 2 – Supplier Collaboration:** Launch supplier portal, contract repository, compliance automation.
4. **Phase 3 – Advanced Analytics:** Deploy ML-driven insights, optimization recommendations, predictive alerts.
5. **Phase 4 – Continuous Improvement:** Expand integrations, refine mobile experience, iterate on KPIs, and roll out to additional business units.

## Risks & Assumptions

- Assumes ERP exposes modern APIs or middleware layer; otherwise, integration timeline extends.
- Data quality in legacy supplier and contract records may require cleansing effort.
- Change management and training are critical to drive adoption among site buyers accustomed to manual processes.
- Vendor onboarding compliance varies by region; configuration flexibility must handle local nuances without code changes.
- Open-source components selected must have active communities or commercial support options; additional vetting may extend procurement of technology stack.
