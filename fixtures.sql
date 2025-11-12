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

-- Products (linked to brands and specifications)
INSERT INTO products (id, name, brand_id, specification_id, created_at, updated_at) VALUES
(1, 'MacBook Pro 15"', 1, 1, datetime('now'), datetime('now')),
(2, 'Dell XPS 15', 2, 1, datetime('now'), datetime('now')),
(3, 'ThinkPad X1 Carbon', 6, 1, datetime('now'), datetime('now')),
(4, 'Dell UltraSharp U2720Q', 2, 2, datetime('now'), datetime('now')),
(5, 'LG 27UK850-W', 7, 2, datetime('now'), datetime('now')),
(6, 'Samsung Galaxy S24 Ultra', 3, 3, datetime('now'), datetime('now')),
(7, 'iPhone 15 Pro Max', 1, 3, datetime('now'), datetime('now')),
(8, 'Logitech MX Master 3S', 4, 4, datetime('now'), datetime('now')),
(9, 'Logitech MX Mechanical', 4, 5, datetime('now'), datetime('now'));

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

-- Quotes (vendor price quotes for products)
INSERT INTO quotes (id, vendor_id, product_id, price, currency, converted_price, conversion_rate, quote_date, notes, created_at, updated_at) VALUES
(1, 1, 1, 2499.00, 'USD', 2499.00, 1.0, datetime('now'), 'MacBook Pro with educational discount', datetime('now'), datetime('now')),
(2, 2, 1, 2399.00, 'USD', 2399.00, 1.0, datetime('now'), 'Amazon Business bulk pricing', datetime('now'), datetime('now')),
(3, 1, 2, 1599.99, 'USD', 1599.99, 1.0, datetime('now'), 'Dell XPS on sale', datetime('now'), datetime('now')),
(4, 4, 2, 1699.00, 'USD', 1699.00, 1.0, datetime('now'), 'CDW corporate pricing', datetime('now'), datetime('now')),
(5, 3, 4, 649.99, 'USD', 649.99, 1.0, datetime('now'), 'Monitor with free shipping', datetime('now'), datetime('now')),
(6, 2, 5, 599.00, 'USD', 599.00, 1.0, datetime('now'), 'Amazon Prime discount', datetime('now'), datetime('now')),
(7, 1, 7, 1199.00, 'USD', 1199.00, 1.0, datetime('now'), 'iPhone with trade-in credit', datetime('now'), datetime('now')),
(8, 5, 6, 5999.00, 'CNY', 839.86, 0.14, datetime('now'), 'Bulk order available', datetime('now'), datetime('now')),
(9, 2, 8, 99.99, 'USD', 99.99, 1.0, datetime('now'), 'Logitech mouse with Prime', datetime('now'), datetime('now')),
(10, 4, 9, 149.00, 'USD', 149.00, 1.0, datetime('now'), 'Mechanical keyboard corporate', datetime('now'), datetime('now'));

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
