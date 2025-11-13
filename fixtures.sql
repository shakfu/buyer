-- Buyer Application Fixtures
-- Sample data for development and testing

-- Specifications (general product types)
INSERT INTO specifications (id, name, description, created_at, updated_at) VALUES
(1, 'Laptop - 15 inch', 'Standard 15-inch laptop for general office work', datetime('now'), datetime('now')),
(2, 'Monitor - 27 inch 4K', '27-inch 4K resolution monitor', datetime('now'), datetime('now')),
(3, 'Smartphone - Flagship', 'High-end flagship smartphone', datetime('now'), datetime('now')),
(4, 'Wireless Mouse', 'Ergonomic wireless mouse', datetime('now'), datetime('now')),
(5, 'Mechanical Keyboard', 'Mechanical keyboard with RGB backlight', datetime('now'), datetime('now'));

-- Brands
INSERT INTO brands (id, name, created_at, updated_at) VALUES
(1, 'Apple', datetime('now'), datetime('now')),
(2, 'Dell', datetime('now'), datetime('now')),
(3, 'Samsung', datetime('now'), datetime('now')),
(4, 'Logitech', datetime('now'), datetime('now')),
(5, 'HP', datetime('now'), datetime('now')),
(6, 'Lenovo', datetime('now'), datetime('now')),
(7, 'LG', datetime('now'), datetime('now'));

-- Products (linked to brands and specifications with extended details)
INSERT INTO products (id, name, sku, description, brand_id, specification_id, unit_of_measure, min_order_qty, lead_time_days, is_active, created_by, created_at, updated_at) VALUES
(1, 'MacBook Pro 15"', 'APPLE-MBP15-2024', 'High-performance laptop with M3 Pro chip, 15-inch Liquid Retina display', 1, 1, 'each', 1, 7, 1, 'admin', datetime('now'), datetime('now')),
(2, 'Dell XPS 15', 'DELL-XPS15-9530', 'Premium Windows laptop with Intel Core i7, 15.6-inch OLED display', 2, 1, 'each', 1, 5, 1, 'admin', datetime('now'), datetime('now')),
(3, 'ThinkPad X1 Carbon', 'LENOVO-X1C-G11', 'Business ultrabook with Intel Core i5, 14-inch display', 6, 1, 'each', 5, 10, 1, 'admin', datetime('now'), datetime('now')),
(4, 'Dell UltraSharp U2720Q', 'DELL-U2720Q', 'Professional 27-inch 4K IPS monitor with USB-C hub', 2, 2, 'each', 2, 7, 1, 'admin', datetime('now'), datetime('now')),
(5, 'LG 27UK850-W', 'LG-27UK850', '27-inch 4K UHD IPS monitor with HDR10 support', 7, 2, 'each', 1, 5, 1, 'admin', datetime('now'), datetime('now')),
(6, 'Samsung Galaxy S24 Ultra', 'SAMS-S24U-256', 'Flagship Android smartphone with 256GB storage, S Pen', 3, 3, 'each', 1, 3, 1, 'admin', datetime('now'), datetime('now')),
(7, 'iPhone 15 Pro Max', 'APPLE-IP15PM-256', 'Premium iPhone with A17 Pro chip, 256GB storage, titanium design', 1, 3, 'each', 1, 2, 1, 'admin', datetime('now'), datetime('now')),
(8, 'Logitech MX Master 3S', 'LOGI-MXM3S-BLK', 'Ergonomic wireless mouse with 8K DPI sensor and quiet clicks', 4, 4, 'each', 10, 3, 1, 'admin', datetime('now'), datetime('now')),
(9, 'Logitech MX Mechanical', 'LOGI-MXMECH-BRN', 'Wireless mechanical keyboard with tactile brown switches and backlight', 4, 5, 'each', 5, 3, 1, 'admin', datetime('now'), datetime('now'));

-- Vendors (with currency and discount codes)
INSERT INTO vendors (id, name, currency, discount_code, created_at, updated_at) VALUES
(1, 'Best Buy', 'USD', 'CORP2024', datetime('now'), datetime('now')),
(2, 'Amazon Business', 'USD', 'BIZPRIME', datetime('now'), datetime('now')),
(3, 'B&H Photo Video', 'USD', '', datetime('now'), datetime('now')),
(4, 'CDW', 'USD', 'ENTERPRISE', datetime('now'), datetime('now')),
(5, 'Alibaba Global', 'CNY', 'BULK20', datetime('now'), datetime('now'));

-- Forex rates (for currency conversion)
INSERT INTO forex (id, from_currency, to_currency, rate, effective_date, created_at, updated_at) VALUES
(1, 'USD', 'USD', 1.0, datetime('now'), datetime('now'), datetime('now')),
(2, 'EUR', 'USD', 1.08, datetime('now'), datetime('now'), datetime('now')),
(3, 'GBP', 'USD', 1.27, datetime('now'), datetime('now'), datetime('now')),
(4, 'CNY', 'USD', 0.14, datetime('now'), datetime('now'), datetime('now')),
(5, 'JPY', 'USD', 0.0067, datetime('now'), datetime('now'), datetime('now'));

-- Quotes (vendor price quotes for products with versioning and status)
INSERT INTO quotes (id, vendor_id, product_id, version, price, currency, converted_price, conversion_rate, min_quantity, status, quote_date, notes, created_by, created_at, updated_at) VALUES
-- Active quotes (current versions)
(1, 1, 1, 2, 2499.00, 'USD', 2499.00, 1.0, 1, 'active', datetime('now'), 'MacBook Pro with educational discount - updated price', 'sales_team', datetime('now'), datetime('now')),
(2, 2, 1, 1, 2399.00, 'USD', 2399.00, 1.0, 3, 'active', datetime('now'), 'Amazon Business bulk pricing - minimum 3 units', 'procurement', datetime('now'), datetime('now')),
(3, 1, 2, 1, 1599.99, 'USD', 1599.99, 1.0, 1, 'superseded', datetime('now', '-7 days'), 'Dell XPS on sale - OLD QUOTE', 'sales_team', datetime('now', '-7 days'), datetime('now')),
(4, 4, 2, 2, 1699.00, 'USD', 1699.00, 1.0, 5, 'active', datetime('now'), 'CDW corporate pricing - minimum 5 units for discount', 'procurement', datetime('now'), datetime('now')),
(5, 3, 4, 1, 649.99, 'USD', 649.99, 1.0, 2, 'accepted', datetime('now', '-5 days'), 'Monitor with free shipping - ACCEPTED', 'manager', datetime('now', '-5 days'), datetime('now', '-5 days')),
(6, 2, 5, 1, 599.00, 'USD', 599.00, 1.0, 1, 'active', datetime('now'), 'Amazon Prime discount', 'procurement', datetime('now'), datetime('now')),
(7, 1, 7, 1, 1199.00, 'USD', 1199.00, 1.0, 1, 'active', datetime('now'), 'iPhone with trade-in credit', 'sales_team', datetime('now'), datetime('now')),
(8, 5, 6, 1, 5999.00, 'CNY', 839.86, 0.14, 10, 'active', datetime('now'), 'Bulk order available - minimum 10 units', 'procurement', datetime('now'), datetime('now')),
(9, 2, 8, 1, 99.99, 'USD', 99.99, 1.0, 10, 'active', datetime('now'), 'Logitech mouse with Prime - bulk discount', 'procurement', datetime('now'), datetime('now')),
(10, 4, 9, 1, 149.00, 'USD', 149.00, 1.0, 5, 'active', datetime('now'), 'Mechanical keyboard corporate pricing', 'procurement', datetime('now'), datetime('now')),
-- Superseded quote (replaced by quote 1)
(11, 1, 1, 1, 2599.00, 'USD', 2599.00, 1.0, 1, 'superseded', datetime('now', '-14 days'), 'MacBook Pro - ORIGINAL QUOTE - superseded by better price', 'sales_team', datetime('now', '-14 days'), datetime('now')),
-- Declined quote
(12, 3, 2, 1, 1899.00, 'USD', 1899.00, 1.0, 1, 'declined', datetime('now', '-3 days'), 'Dell XPS - too expensive, declined', 'manager', datetime('now', '-3 days'), datetime('now', '-3 days'));

-- Requisitions (purchasing requirements)
INSERT INTO requisitions (id, name, justification, budget, created_at, updated_at) VALUES
(1, 'Q1 2025 Developer Equipment', 'New hires starting in Q1 need laptops and monitors', 15000.00, datetime('now'), datetime('now')),
(2, 'Sales Team Phone Upgrade', 'Current phones are 3 years old and need replacement', 8000.00, datetime('now'), datetime('now')),
(3, 'Office Peripherals Refresh', 'Replace old mice and keyboards for better ergonomics', 2000.00, datetime('now'), datetime('now'));

-- Requisition Items (line items for requisitions)
INSERT INTO requisition_items (id, requisition_id, specification_id, quantity, budget_per_unit, description, created_at, updated_at) VALUES
(1, 1, 1, 3, 2000.00, 'For new backend developers', datetime('now'), datetime('now')),
(2, 1, 2, 3, 700.00, 'Dual monitor setup for each developer', datetime('now'), datetime('now')),
(3, 2, 3, 5, 1200.00, 'Flagship phones with 5G support', datetime('now'), datetime('now')),
(4, 3, 4, 10, 100.00, 'Ergonomic mice for all staff', datetime('now'), datetime('now')),
(5, 3, 5, 10, 150.00, 'Mechanical keyboards for programmers', datetime('now'), datetime('now'));

-- Projects (with budgets and deadlines)
INSERT INTO projects (id, name, description, budget, deadline, status, created_at, updated_at) VALUES
(1, 'Office Renovation 2025', 'Complete renovation of headquarters office space', 250000.00, datetime('2025-06-30'), 'planning', datetime('now'), datetime('now')),
(2, 'Remote Work Infrastructure', 'Setup equipment for 50 remote employees', 125000.00, datetime('2025-03-31'), 'active', datetime('now'), datetime('now')),
(3, 'Data Center Upgrade', 'Hardware refresh for main data center', 500000.00, datetime('2025-12-31'), 'planning', datetime('now'), datetime('now'));

-- Bill of Materials (one per project)
INSERT INTO bills_of_materials (id, project_id, notes, created_at, updated_at) VALUES
(1, 1, 'Master BOM for office renovation project', datetime('now'), datetime('now')),
(2, 2, 'Equipment list for remote workers', datetime('now'), datetime('now')),
(3, 3, 'Data center hardware specifications', datetime('now'), datetime('now'));

-- Bill of Materials Items (specifications needed for each project)
-- Note: Currently specification_id has a global unique constraint (bug in model - should be composite unique)
-- TODO: Fix BillOfMaterialsItem model to use composite unique index (bill_of_materials_id, specification_id)
-- For now, each specification can only appear in one BOM
INSERT INTO bill_of_materials_items (id, bill_of_materials_id, specification_id, quantity, notes, created_at, updated_at) VALUES
(1, 1, 1, 25, 'Laptops for office staff', datetime('now'), datetime('now')),
(2, 1, 2, 50, 'Dual monitors for each workstation', datetime('now'), datetime('now')),
(3, 1, 4, 30, 'Wireless mice for all desks', datetime('now'), datetime('now')),
(4, 1, 5, 25, 'Mechanical keyboards for developers', datetime('now'), datetime('now')),
(5, 2, 3, 5, 'Smartphones for remote team leads', datetime('now'), datetime('now'));

-- Project Requisitions (created from project BOM items)
INSERT INTO project_requisitions (id, project_id, name, justification, budget, created_at, updated_at) VALUES
(1, 1, 'Phase 1 Office Equipment', 'Initial procurement for office renovation - workstations and peripherals', 75000.00, datetime('now'), datetime('now')),
(2, 2, 'Remote Worker Laptop Batch 1', 'First batch of 10 laptops for remote team expansion', 25000.00, datetime('now'), datetime('now'));

-- Project Requisition Items (specific quantities from BOM items)
INSERT INTO project_requisition_items (id, project_requisition_id, bill_of_materials_item_id, quantity_requested, notes, created_at, updated_at) VALUES
(1, 1, 1, 10, 'First 10 laptops for office staff', datetime('now'), datetime('now')),
(2, 1, 2, 20, 'Monitors for first 10 workstations', datetime('now'), datetime('now')),
(3, 1, 3, 10, 'Mice for first batch', datetime('now'), datetime('now')),
(4, 2, 5, 3, 'Smartphones for 3 remote team leads', datetime('now'), datetime('now'));

-- Purchase Orders (tracking quote acceptance through delivery)
INSERT INTO purchase_orders (id, quote_id, vendor_id, product_id, requisition_id, po_number, status, order_date, expected_delivery, actual_delivery, quantity, unit_price, currency, total_amount, shipping_cost, tax, grand_total, invoice_number, notes, created_at, updated_at) VALUES
-- Approved orders from Q1 Developer Equipment requisition
(1, 2, 2, 1, 1, 'PO-2025-001', 'approved', datetime('2025-01-15'), datetime('2025-02-01'), NULL, 3, 2399.00, 'USD', 7197.00, 150.00, 575.76, 7922.76, NULL, 'MacBook Pro for new developers - Amazon Business pricing', datetime('now'), datetime('now')),

-- Shipped monitor order
(2, 5, 3, 4, 1, 'PO-2025-002', 'shipped', datetime('2025-01-18'), datetime('2025-02-05'), NULL, 3, 649.99, 'USD', 1949.97, 75.00, 155.99, 2180.96, 'INV-BH-45678', 'Dell UltraSharp monitors for developer workstations', datetime('now'), datetime('now')),

-- Received keyboard order
(3, 10, 4, 9, 3, 'PO-2025-003', 'received', datetime('2025-01-10'), datetime('2025-01-25'), datetime('2025-01-24'), 10, 149.00, 'USD', 1490.00, 50.00, 119.20, 1659.20, 'INV-CDW-99234', 'Mechanical keyboards for programmers', datetime('now'), datetime('now')),

-- Pending phone order
(4, 7, 1, 7, 2, 'PO-2025-004', 'pending', datetime('2025-01-20'), datetime('2025-02-10'), NULL, 5, 1199.00, 'USD', 5995.00, 0.00, 479.60, 6474.60, NULL, 'iPhone 15 Pro Max for sales team - awaiting approval', datetime('now'), datetime('now')),

-- Ordered mice from Amazon
(5, 9, 2, 8, 3, 'PO-2025-005', 'ordered', datetime('2025-01-16'), datetime('2025-01-30'), NULL, 10, 99.99, 'USD', 999.90, 0.00, 79.99, 1079.89, NULL, 'Logitech MX Master 3S mice with Prime shipping', datetime('now'), datetime('now')),

-- Cancelled order (changed requirements)
(6, 3, 1, 2, NULL, 'PO-2025-006', 'cancelled', datetime('2025-01-12'), datetime('2025-02-01'), NULL, 2, 1599.99, 'USD', 3199.98, 100.00, 255.99, 3555.97, NULL, 'Cancelled - switched to Lenovo ThinkPads instead', datetime('now'), datetime('now'));

-- Documents (file attachments for various entities)
INSERT INTO documents (id, entity_type, entity_id, file_name, file_type, file_size, file_path, description, uploaded_by, created_at) VALUES
-- Vendor documents
(1, 'vendor', 1, 'best-buy-w9-2024.pdf', 'pdf', 245760, '/docs/vendors/best-buy-w9-2024.pdf', 'W-9 Tax Form for Best Buy', 'admin', datetime('now', '-30 days')),
(2, 'vendor', 2, 'amazon-business-contract.pdf', 'pdf', 512000, '/docs/vendors/amazon-business-contract.pdf', 'Corporate purchasing agreement with Amazon Business', 'procurement', datetime('now', '-60 days')),
(3, 'vendor', 4, 'cdw-pricing-sheet-2025.xlsx', 'xlsx', 89600, '/docs/vendors/cdw-pricing-sheet-2025.xlsx', 'Volume pricing tiers for 2025', 'procurement', datetime('now', '-15 days')),

-- Quote documents
(4, 'quote', 1, 'apple-edu-discount-proof.pdf', 'pdf', 128000, '/docs/quotes/apple-edu-discount-proof.pdf', 'Educational institution verification for discount', 'sales_team', datetime('now', '-5 days')),
(5, 'quote', 5, 'monitor-specs-sheet.pdf', 'pdf', 342400, '/docs/quotes/monitor-specs-sheet.pdf', 'Detailed technical specifications', 'procurement', datetime('now', '-6 days')),
(6, 'quote', 8, 'alibaba-bulk-pricing.pdf', 'pdf', 156800, '/docs/quotes/alibaba-bulk-pricing.pdf', 'Bulk pricing tiers for quantities over 100 units', 'procurement', datetime('now', '-2 days')),

-- Purchase Order documents
(7, 'purchase_order', 2, 'shipping-label-PO-2025-002.pdf', 'pdf', 67200, '/docs/pos/shipping-label-PO-2025-002.pdf', 'B&H Photo shipping label and tracking', 'logistics', datetime('now', '-3 days')),
(8, 'purchase_order', 3, 'invoice-INV-CDW-99234.pdf', 'pdf', 198400, '/docs/pos/invoice-INV-CDW-99234.pdf', 'Final invoice from CDW for keyboards', 'accounting', datetime('now', '-5 days')),
(9, 'purchase_order', 3, 'delivery-receipt-signed.jpg', 'jpg', 2048000, '/docs/pos/delivery-receipt-signed.jpg', 'Signed delivery receipt with timestamp', 'warehouse', datetime('now', '-6 days')),

-- Product documents
(10, 'product', 1, 'macbook-pro-datasheet.pdf', 'pdf', 445440, '/docs/products/macbook-pro-datasheet.pdf', 'Official Apple technical specifications', 'admin', datetime('now', '-20 days')),
(11, 'product', 4, 'dell-monitor-warranty.pdf', 'pdf', 234560, '/docs/products/dell-monitor-warranty.pdf', '3-year premium warranty details', 'admin', datetime('now', '-18 days')),
(12, 'product', 9, 'logitech-keyboard-manual.pdf', 'pdf', 678400, '/docs/products/logitech-keyboard-manual.pdf', 'User manual and quick start guide', 'admin', datetime('now', '-10 days'));
