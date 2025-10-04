# Operations Workbook

The `buyer report` command exports a multi-sheet Excel workbook at the path you provide via `--output` (or the `BUYER_REPORT_PATH` environment variable). The workbook is assembled with the in-tree XLSX builder and reflects the live data managed by the repository backing the CLI.

## Sheets

- **Summary** – key totals (active suppliers, open POs, upcoming deliveries, pending approvals, invoices on hold) and the top five alerts plus leading suppliers by open PO value.
- **Suppliers** – a tabular export of the supplier master, including country, category, risk rating, activity status, and spend year-to-date.
- **Purchase Orders** – every purchase order with currency, total value, and expected delivery for downstream scheduling reconciliation.
- **Approvals** – the approval queue with requester, current approver, due date, and status for gating compliance reviews.
- **Deliveries** – inbound milestones with expected dates and status, enabling quick escalation of delayed logistics.
- **Invoices** – current invoice positions with amount, due date, and hold status for Accounts Payable follow-up.

## Usage Tips

1. Bootstrap fixtures with `swift run buyer fixtures --store sqlite --output fixtures/procurement.sqlite3` (or `--store json` if you prefer a ledger).
2. Set `BUYER_DB_PATH` to keep the data store in a persistent location (e.g., `~/.buyer/procurement.sqlite3`). Use a `.json` extension if you prefer the file-backed ledger.
3. Pass `--output reports/ops.xlsx` to write the workbook to a tracked folder (`reports/` is created automatically).
4. Re-run `buyer report` before governance meetings to refresh the workbook with the latest approvals, deliveries, and invoice states.

For extensions, add new sheets or metrics inside `writeProcurementWorkbook` in `src/buyerlib/commands/xlsxwriter.swift` and expose new data from `ProcurementService` so the CLI stays thin.
