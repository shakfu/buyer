# Procurement Management System Data Model

## Overview
This data model captures the core entities required to manage procurement within a large construction enterprise. Each entity description below lists its primary purpose, critical attributes, relationships, and notable constraints needed to preserve data quality and ensure traceability across the procurement lifecycle.

## Entity Catalogue
- **Project** – Represents a construction program or sub-project aligned to SAP projects/WBS elements.
- **ProcurementDemand** – Captures specific material or service requirements raised against a project.
- **Supplier** – Stores vendor master data synchronized with SAP.
- **SupplierQualification** – Tracks compliance artefacts, certifications, and approval checkpoints for suppliers.
- **RFxEvent** – Defines sourcing events (RFI/RFQ/RFP/e-auction) tied to project demand.
- **RFxLineItem** – Individual lines within an RFx aligned to demands and scope packages.
- **RFxResponse** – Supplier submissions with commercial and technical data.
- **RFxScorecard** – Evaluation records per response and evaluator.
- **Contract** – Awarded agreements (frame, blanket, or project-specific) derived from sourcing outcomes.
- **PurchaseOrder** – Formal orders issued to suppliers; optionally bound to contracts.
- **PurchaseOrderLine** – Item-level commitments, quantities, and delivery schedules.
- **ReleaseOrder** – Drawdowns against frame/blanket contracts or overarching POs.
- **DeliveryMilestone** – Planned or actual delivery checkpoints for each PO line or release.
- **InventoryLocation** – Warehouses, laydown yards, or site stores receiving materials.
- **GoodsReceipt** – Evidence of goods/services receipt with quantity and quality status.
- **InspectionReport** – Quality inspection outcomes linked to receipts.
- **NonConformance** – Recorded issues requiring corrective action.
- **Invoice** – Supplier invoice header information for 3-way matching.
- **InvoiceLine** – Line-level invoice details mapped to PO lines or receipts.
- **PaymentStatusHistory** – Chronological payment updates from SAP.
- **UserAccount** – Represents people or service accounts interacting with the PMS.
- **Role** – Role definitions used for authorization.
- **RoleMembership** – Associates users with roles and optional project scope.
- **ApprovalPolicy** – Configurable approval matrices per object type and spend threshold.
- **ApprovalPolicyStep** – Ordered steps within a policy including routing rules.
- **ApprovalRequest** – Runtime approval instances awaiting or recording decisions.
- **AuditLog** – Immutable log of user or system actions.

## Entity Details & Constraints
### Project
- **Key Attributes:** `id`, `tenant_id`, `project_code`, `sap_wbs_element`, `status`, `start_date`, `end_date`.
- **Relationships:** One-to-many with ProcurementDemand, Contract, PurchaseOrder, and Invoice. Linked to RoleMembership for project-specific access.
- **Constraints:** `project_code` unique per tenant; `project_code` must match SAP reference. Lifecycle dates cannot overlap in conflict with parent project (enforced via business rules).

### ProcurementDemand
- **Key Attributes:** `id`, `tenant_id`, `project_id`, `category`, `description`, `required_date`, `quantity`, `uom`, `status`.
- **Relationships:** Many-to-one with Project; one-to-many with RFxLineItem and PurchaseOrderLine.
- **Constraints:** `required_date` ≥ Project `start_date`; status transitions governed by workflow engine. Quantity must be positive.

### Supplier
- **Key Attributes:** `id`, `tenant_id`, `sap_vendor_id`, `name`, `country`, `status`, `risk_rating`.
- **Relationships:** One-to-many with SupplierQualification, RFxResponse, Contract, PurchaseOrder, and Invoice.
- **Constraints:** `sap_vendor_id` unique per tenant; only active suppliers may receive POs.

### SupplierQualification
- **Key Attributes:** `id`, `supplier_id`, `qualification_type`, `valid_from`, `valid_to`, `status`, `document_uri`.
- **Relationships:** Many-to-one with Supplier.
- **Constraints:** Validity windows cannot overlap for the same type; `status` must be compliant before supplier is set to active.

### RFxEvent
- **Key Attributes:** `id`, `tenant_id`, `project_id`, `event_number`, `type`, `status`, `submission_deadline`.
- **Relationships:** One-to-many with RFxLineItem and RFxResponse; linked to ApprovalRequest for releases.
- **Constraints:** `event_number` unique per tenant; deadline must be in the future when publishing; status progression controlled by sourcing workflow.

### RFxLineItem
- **Key Attributes:** `id`, `rfx_event_id`, `demand_id`, `item_number`, `description`, `quantity`, `uom`, `evaluation_weight`.
- **Relationships:** Many-to-one with RFxEvent; optional link to ProcurementDemand.
- **Constraints:** `item_number` unique within RFxEvent; weights must total 100% per event when using weighted scoring.

### RFxResponse
- **Key Attributes:** `id`, `rfx_event_id`, `supplier_id`, `submitted_at`, `commercial_score`, `technical_score`, `currency`.
- **Relationships:** Many-to-one with RFxEvent and Supplier; one-to-many with RFxScorecard.
- **Constraints:** A supplier can only submit one active response per event; scores derived from scorecards.

### RFxScorecard
- **Key Attributes:** `id`, `rfx_response_id`, `evaluator_id`, `criterion`, `score`, `comments`.
- **Relationships:** Many-to-one with RFxResponse; references UserAccount as evaluator.
- **Constraints:** Score range (0-100) enforced; each evaluator may score each criterion once per response.

### Contract
- **Key Attributes:** `id`, `tenant_id`, `contract_number`, `supplier_id`, `project_id`, `rfx_event_id`, `type`, `effective_date`, `expiry_date`, `status`.
- **Relationships:** Many-to-one with Supplier and Project; optional link to RFxEvent; one-to-many with PurchaseOrder and ReleaseOrder.
- **Constraints:** `contract_number` unique per tenant; expiry must be after effective date; active contracts require valid supplier qualifications.

### PurchaseOrder
- **Key Attributes:** `id`, `tenant_id`, `po_number`, `supplier_id`, `project_id`, `contract_id`, `issued_at`, `currency`, `status`, `sap_po_number`.
- **Relationships:** Many-to-one with Supplier, Project, Contract; one-to-many with PurchaseOrderLine, ReleaseOrder, GoodsReceipt, InvoiceLine.
- **Constraints:** `po_number` unique per tenant; `sap_po_number` unique when synced; cannot issue unless contract (when required) is active.

### PurchaseOrderLine
- **Key Attributes:** `id`, `purchase_order_id`, `demand_id`, `line_number`, `description`, `quantity`, `uom`, `unit_price`, `delivery_start`, `delivery_end`, `status`.
- **Relationships:** Many-to-one with PurchaseOrder; optional link to ProcurementDemand; one-to-many with DeliveryMilestone, GoodsReceipt, InvoiceLine.
- **Constraints:** `line_number` unique within PurchaseOrder; delivery windows must fall within project timeline; quantity balance validated against receipts and invoices.

### ReleaseOrder
- **Key Attributes:** `id`, `purchase_order_id`, `contract_id`, `release_number`, `line_reference_id`, `quantity`, `release_date`, `status`.
- **Relationships:** Many-to-one with PurchaseOrder and Contract; optionally tied to a specific PurchaseOrderLine through `line_reference_id`.
- **Constraints:** `release_number` unique per contract; quantity must not exceed available balance on referenced line or contract ceiling.

### DeliveryMilestone
- **Key Attributes:** `id`, `purchase_order_line_id`, `expected_date`, `expected_quantity`, `actual_date`, `actual_quantity`, `status`.
- **Relationships:** Many-to-one with PurchaseOrderLine.
- **Constraints:** Expected/actual quantities must not exceed line quantity; status driven by goods receipt completion.

### InventoryLocation
- **Key Attributes:** `id`, `tenant_id`, `code`, `name`, `site_type`, `project_id`, `address`.
- **Relationships:** Optional link to Project; one-to-many with GoodsReceipt.
- **Constraints:** `code` unique per tenant; if linked to project, must belong to same tenant.

### GoodsReceipt
- **Key Attributes:** `id`, `purchase_order_line_id`, `release_order_id`, `inventory_location_id`, `grn_number`, `received_date`, `received_quantity`, `accepted_quantity`, `received_by`, `status`.
- **Relationships:** Many-to-one with PurchaseOrderLine and InventoryLocation; optional link to ReleaseOrder; one-to-many with InspectionReport and NonConformance; referenced by InvoiceLine.
- **Constraints:** `received_quantity` ≥ `accepted_quantity`; totals cannot exceed ordered balance; GRN number unique per tenant.

### InspectionReport
- **Key Attributes:** `id`, `goods_receipt_id`, `inspector_id`, `inspection_date`, `result`, `remarks`.
- **Relationships:** Many-to-one with GoodsReceipt; references UserAccount as inspector.
- **Constraints:** Result constrained to enumerated values (Pass/Conditional/Fail); inspection date ≥ receipt date.

### NonConformance
- **Key Attributes:** `id`, `goods_receipt_id`, `reported_date`, `severity`, `issue_type`, `quantity_affected`, `status`, `resolution_notes`.
- **Relationships:** Many-to-one with GoodsReceipt; may drive ApprovalRequest for concessions.
- **Constraints:** Severity enumerations; quantity affected ≤ accepted quantity; cannot close until linked corrective actions complete.

### Invoice
- **Key Attributes:** `id`, `tenant_id`, `invoice_number`, `supplier_id`, `project_id`, `invoice_date`, `due_date`, `currency`, `total_amount`, `status`, `sap_invoice_id`.
- **Relationships:** Many-to-one with Supplier and Project; one-to-many with InvoiceLine and PaymentStatusHistory.
- **Constraints:** `invoice_number` unique per supplier per tenant; due date ≥ invoice date; status transitions controlled by matching process.

### InvoiceLine
- **Key Attributes:** `id`, `invoice_id`, `purchase_order_line_id`, `goods_receipt_id`, `line_number`, `description`, `quantity`, `unit_price`, `amount`.
- **Relationships:** Many-to-one with Invoice; optional links to PurchaseOrderLine and GoodsReceipt.
- **Constraints:** `amount` = `quantity` × `unit_price`; quantity cannot exceed receipted balance.

### PaymentStatusHistory
- **Key Attributes:** `id`, `invoice_id`, `status`, `status_date`, `sap_reference`, `notes`.
- **Relationships:** Many-to-one with Invoice.
- **Constraints:** Status values synchronized with SAP (e.g., Parked, Posted, Paid); chronological order enforced by trigger or application logic.

### UserAccount
- **Key Attributes:** `id`, `tenant_id`, `username`, `email`, `status`, `source_system`.
- **Relationships:** Many-to-many with Role via RoleMembership; referenced by RFxScorecard, InspectionReport, ApprovalRequest, AuditLog.
- **Constraints:** `username` unique per tenant; only active users may take approvals or submit transactions.

### Role
- **Key Attributes:** `id`, `tenant_id`, `code`, `name`, `description`.
- **Relationships:** Many-to-many with UserAccount via RoleMembership; referenced in ApprovalPolicyStep.
- **Constraints:** `code` unique per tenant.

### RoleMembership
- **Key Attributes:** `id`, `user_id`, `role_id`, `project_id`, `assigned_at`, `expires_at`.
- **Relationships:** Many-to-one with UserAccount, Role, optional Project.
- **Constraints:** Membership must align to tenant; optional expiry enforces temporary assignments; duplicates prevented by unique constraint (`user_id`, `role_id`, `project_id`).

### ApprovalPolicy
- **Key Attributes:** `id`, `tenant_id`, `object_type`, `threshold_currency`, `threshold_amount`, `active_from`, `active_to`, `status`.
- **Relationships:** One-to-many with ApprovalPolicyStep; referenced by ApprovalRequest.
- **Constraints:** Only one active policy per object type and threshold range per tenant; validity windows cannot overlap.

### ApprovalPolicyStep
- **Key Attributes:** `id`, `approval_policy_id`, `sequence`, `role_id`, `rule_expression`, `sla_hours`.
- **Relationships:** Many-to-one with ApprovalPolicy and Role.
- **Constraints:** Sequence must be continuous starting at 1; SLA positive.

### ApprovalRequest
- **Key Attributes:** `id`, `tenant_id`, `object_type`, `object_id`, `policy_id`, `current_step_id`, `status`, `created_at`, `completed_at`.
- **Relationships:** Many-to-one with ApprovalPolicy and ApprovalPolicyStep; one-to-many with AuditLog entries; references UserAccount for approver decisions (captured as audit entries).
- **Constraints:** Combination (`object_type`, `object_id`) must be unique for open approvals; status transitions managed by workflow engine.

### AuditLog
- **Key Attributes:** `id`, `tenant_id`, `actor_id`, `action`, `object_type`, `object_id`, `details`, `created_at`, `source_ip`.
- **Relationships:** References UserAccount for actor; optional link to ApprovalRequest or other entities via polymorphic keys.
- **Constraints:** Immutable records; `created_at` default current timestamp; `details` stored as JSON with schema validation.

## Prototype SQLite Schema
```sql
PRAGMA foreign_keys = ON;

CREATE TABLE project (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    project_code TEXT NOT NULL,
    sap_wbs_element TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, project_code)
);

CREATE TABLE procurement_demand (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    category TEXT NOT NULL,
    description TEXT NOT NULL,
    required_date DATE NOT NULL,
    quantity REAL NOT NULL CHECK (quantity > 0),
    uom TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES project(id)
);

CREATE TABLE supplier (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    sap_vendor_id TEXT NOT NULL,
    name TEXT NOT NULL,
    country TEXT,
    status TEXT NOT NULL,
    risk_rating TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, sap_vendor_id)
);

CREATE TABLE supplier_qualification (
    id TEXT PRIMARY KEY,
    supplier_id TEXT NOT NULL,
    qualification_type TEXT NOT NULL,
    valid_from DATE NOT NULL,
    valid_to DATE,
    status TEXT NOT NULL,
    document_uri TEXT,
    FOREIGN KEY (supplier_id) REFERENCES supplier(id)
);

CREATE TABLE rfx_event (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    event_number TEXT NOT NULL,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    submission_deadline DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, event_number),
    FOREIGN KEY (project_id) REFERENCES project(id)
);

CREATE TABLE rfx_line_item (
    id TEXT PRIMARY KEY,
    rfx_event_id TEXT NOT NULL,
    demand_id TEXT,
    item_number INTEGER NOT NULL,
    description TEXT NOT NULL,
    quantity REAL NOT NULL CHECK (quantity > 0),
    uom TEXT NOT NULL,
    evaluation_weight REAL,
    FOREIGN KEY (rfx_event_id) REFERENCES rfx_event(id),
    FOREIGN KEY (demand_id) REFERENCES procurement_demand(id),
    UNIQUE (rfx_event_id, item_number)
);

CREATE TABLE rfx_response (
    id TEXT PRIMARY KEY,
    rfx_event_id TEXT NOT NULL,
    supplier_id TEXT NOT NULL,
    submitted_at DATETIME,
    commercial_score REAL,
    technical_score REAL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    FOREIGN KEY (rfx_event_id) REFERENCES rfx_event(id),
    FOREIGN KEY (supplier_id) REFERENCES supplier(id),
    UNIQUE (rfx_event_id, supplier_id)
);

CREATE TABLE user_account (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    status TEXT NOT NULL,
    source_system TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, username)
);

CREATE TABLE rfx_scorecard (
    id TEXT PRIMARY KEY,
    rfx_response_id TEXT NOT NULL,
    evaluator_id TEXT NOT NULL,
    criterion TEXT NOT NULL,
    score REAL NOT NULL CHECK (score BETWEEN 0 AND 100),
    comments TEXT,
    UNIQUE (rfx_response_id, evaluator_id, criterion),
    FOREIGN KEY (rfx_response_id) REFERENCES rfx_response(id),
    FOREIGN KEY (evaluator_id) REFERENCES user_account(id)
);

CREATE TABLE contract (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    contract_number TEXT NOT NULL,
    supplier_id TEXT NOT NULL,
    project_id TEXT,
    rfx_event_id TEXT,
    type TEXT NOT NULL,
    effective_date DATE NOT NULL,
    expiry_date DATE,
    status TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, contract_number),
    FOREIGN KEY (supplier_id) REFERENCES supplier(id),
    FOREIGN KEY (project_id) REFERENCES project(id),
    FOREIGN KEY (rfx_event_id) REFERENCES rfx_event(id)
);

CREATE TABLE purchase_order (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    po_number TEXT NOT NULL,
    supplier_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    contract_id TEXT,
    issued_at DATETIME NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    sap_po_number TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, po_number),
    UNIQUE (tenant_id, sap_po_number),
    FOREIGN KEY (supplier_id) REFERENCES supplier(id),
    FOREIGN KEY (project_id) REFERENCES project(id),
    FOREIGN KEY (contract_id) REFERENCES contract(id)
);

CREATE TABLE purchase_order_line (
    id TEXT PRIMARY KEY,
    purchase_order_id TEXT NOT NULL,
    demand_id TEXT,
    line_number INTEGER NOT NULL,
    description TEXT NOT NULL,
    quantity REAL NOT NULL CHECK (quantity > 0),
    uom TEXT NOT NULL,
    unit_price REAL NOT NULL CHECK (unit_price >= 0),
    delivery_start DATE,
    delivery_end DATE,
    status TEXT NOT NULL,
    FOREIGN KEY (purchase_order_id) REFERENCES purchase_order(id),
    FOREIGN KEY (demand_id) REFERENCES procurement_demand(id),
    UNIQUE (purchase_order_id, line_number)
);

CREATE TABLE release_order (
    id TEXT PRIMARY KEY,
    purchase_order_id TEXT NOT NULL,
    contract_id TEXT,
    line_reference_id TEXT,
    release_number TEXT NOT NULL,
    quantity REAL NOT NULL CHECK (quantity > 0),
    release_date DATE NOT NULL,
    status TEXT NOT NULL,
    FOREIGN KEY (purchase_order_id) REFERENCES purchase_order(id),
    FOREIGN KEY (contract_id) REFERENCES contract(id),
    FOREIGN KEY (line_reference_id) REFERENCES purchase_order_line(id),
    UNIQUE (contract_id, release_number)
);

CREATE TABLE delivery_milestone (
    id TEXT PRIMARY KEY,
    purchase_order_line_id TEXT NOT NULL,
    expected_date DATE NOT NULL,
    expected_quantity REAL NOT NULL CHECK (expected_quantity >= 0),
    actual_date DATE,
    actual_quantity REAL CHECK (actual_quantity >= 0),
    status TEXT NOT NULL,
    FOREIGN KEY (purchase_order_line_id) REFERENCES purchase_order_line(id)
);

CREATE TABLE inventory_location (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    site_type TEXT NOT NULL,
    project_id TEXT,
    address TEXT,
    UNIQUE (tenant_id, code),
    FOREIGN KEY (project_id) REFERENCES project(id)
);

CREATE TABLE goods_receipt (
    id TEXT PRIMARY KEY,
    purchase_order_line_id TEXT NOT NULL,
    release_order_id TEXT,
    inventory_location_id TEXT NOT NULL,
    grn_number TEXT NOT NULL,
    received_date DATE NOT NULL,
    received_quantity REAL NOT NULL CHECK (received_quantity >= 0),
    accepted_quantity REAL NOT NULL CHECK (accepted_quantity >= 0),
    received_by TEXT NOT NULL,
    status TEXT NOT NULL,
    UNIQUE (inventory_location_id, grn_number),
    FOREIGN KEY (purchase_order_line_id) REFERENCES purchase_order_line(id),
    FOREIGN KEY (release_order_id) REFERENCES release_order(id),
    FOREIGN KEY (inventory_location_id) REFERENCES inventory_location(id),
    FOREIGN KEY (received_by) REFERENCES user_account(id)
);

CREATE TABLE inspection_report (
    id TEXT PRIMARY KEY,
    goods_receipt_id TEXT NOT NULL,
    inspector_id TEXT NOT NULL,
    inspection_date DATE NOT NULL,
    result TEXT NOT NULL,
    remarks TEXT,
    FOREIGN KEY (goods_receipt_id) REFERENCES goods_receipt(id),
    FOREIGN KEY (inspector_id) REFERENCES user_account(id)
);

CREATE TABLE non_conformance (
    id TEXT PRIMARY KEY,
    goods_receipt_id TEXT NOT NULL,
    reported_date DATE NOT NULL,
    severity TEXT NOT NULL,
    issue_type TEXT NOT NULL,
    quantity_affected REAL NOT NULL CHECK (quantity_affected >= 0),
    status TEXT NOT NULL,
    resolution_notes TEXT,
    FOREIGN KEY (goods_receipt_id) REFERENCES goods_receipt(id)
);

CREATE TABLE invoice (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    invoice_number TEXT NOT NULL,
    supplier_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    currency TEXT NOT NULL,
    total_amount REAL NOT NULL CHECK (total_amount >= 0),
    status TEXT NOT NULL,
    sap_invoice_id TEXT,
    UNIQUE (tenant_id, supplier_id, invoice_number),
    UNIQUE (tenant_id, sap_invoice_id),
    FOREIGN KEY (supplier_id) REFERENCES supplier(id),
    FOREIGN KEY (project_id) REFERENCES project(id)
);

CREATE TABLE invoice_line (
    id TEXT PRIMARY KEY,
    invoice_id TEXT NOT NULL,
    purchase_order_line_id TEXT,
    goods_receipt_id TEXT,
    line_number INTEGER NOT NULL,
    description TEXT NOT NULL,
    quantity REAL NOT NULL CHECK (quantity >= 0),
    unit_price REAL NOT NULL CHECK (unit_price >= 0),
    amount REAL NOT NULL CHECK (amount >= 0),
    FOREIGN KEY (invoice_id) REFERENCES invoice(id),
    FOREIGN KEY (purchase_order_line_id) REFERENCES purchase_order_line(id),
    FOREIGN KEY (goods_receipt_id) REFERENCES goods_receipt(id),
    UNIQUE (invoice_id, line_number)
);

CREATE TABLE payment_status_history (
    id TEXT PRIMARY KEY,
    invoice_id TEXT NOT NULL,
    status TEXT NOT NULL,
    status_date DATETIME NOT NULL,
    sap_reference TEXT,
    notes TEXT,
    FOREIGN KEY (invoice_id) REFERENCES invoice(id)
);

CREATE TABLE role (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    UNIQUE (tenant_id, code)
);

CREATE TABLE role_membership (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    project_id TEXT,
    assigned_at DATETIME NOT NULL,
    expires_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES user_account(id),
    FOREIGN KEY (role_id) REFERENCES role(id),
    FOREIGN KEY (project_id) REFERENCES project(id),
    UNIQUE (user_id, role_id, project_id)
);

CREATE TABLE approval_policy (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    object_type TEXT NOT NULL,
    threshold_currency TEXT,
    threshold_amount REAL,
    active_from DATE NOT NULL,
    active_to DATE,
    status TEXT NOT NULL,
    UNIQUE (tenant_id, object_type, active_from),
    CHECK (threshold_amount IS NULL OR threshold_amount >= 0)
);

CREATE TABLE approval_policy_step (
    id TEXT PRIMARY KEY,
    approval_policy_id TEXT NOT NULL,
    sequence INTEGER NOT NULL,
    role_id TEXT NOT NULL,
    rule_expression TEXT,
    sla_hours INTEGER NOT NULL CHECK (sla_hours > 0),
    FOREIGN KEY (approval_policy_id) REFERENCES approval_policy(id),
    FOREIGN KEY (role_id) REFERENCES role(id),
    UNIQUE (approval_policy_id, sequence)
);

CREATE TABLE approval_request (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    object_type TEXT NOT NULL,
    object_id TEXT NOT NULL,
    policy_id TEXT NOT NULL,
    current_step_id TEXT,
    status TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    completed_at DATETIME,
    FOREIGN KEY (policy_id) REFERENCES approval_policy(id),
    FOREIGN KEY (current_step_id) REFERENCES approval_policy_step(id),
    UNIQUE (object_type, object_id, status) WHERE status IN ('Pending','InProgress')
);

CREATE TABLE audit_log (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    actor_id TEXT,
    action TEXT NOT NULL,
    object_type TEXT NOT NULL,
    object_id TEXT,
    details TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    source_ip TEXT,
    FOREIGN KEY (actor_id) REFERENCES user_account(id)
);
```
