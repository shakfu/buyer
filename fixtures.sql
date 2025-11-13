-- Buyer Application Fixtures
-- Sample data for development and testing

-- Specifications (general product types)
INSERT INTO specifications (id, name, description, created_at, updated_at) VALUES
(1, 'Laptop - 15 inch', 'Standard 15-inch laptop for general office work', datetime('now'), datetime('now')),
(2, 'Monitor - 27 inch 4K', '27-inch 4K resolution monitor', datetime('now'), datetime('now')),
(3, 'Smartphone - Flagship', 'High-end flagship smartphone', datetime('now'), datetime('now')),
(4, 'Wireless Mouse', 'Ergonomic wireless mouse', datetime('now'), datetime('now')),
(5, 'Mechanical Keyboard', 'Mechanical keyboard with RGB backlight', datetime('now'), datetime('now'));

-- Specification Attributes (comparable features for each specification type)
INSERT INTO specification_attributes (id, specification_id, name, data_type, unit, is_required, min_value, max_value, description, created_at, updated_at) VALUES
-- Laptop attributes
(1, 1, 'RAM', 'number', 'GB', 1, 4, 128, 'Memory capacity', datetime('now'), datetime('now')),
(2, 1, 'Storage', 'number', 'GB', 1, 128, 4096, 'Storage capacity', datetime('now'), datetime('now')),
(3, 1, 'Screen Size', 'number', 'inches', 1, 13, 17, 'Display diagonal size', datetime('now'), datetime('now')),
(4, 1, 'CPU Cores', 'number', 'cores', 0, 2, 32, 'Number of CPU cores', datetime('now'), datetime('now')),
(5, 1, 'Storage Type', 'text', NULL, 1, NULL, NULL, 'SSD or HDD', datetime('now'), datetime('now')),
(6, 1, 'Has Touchscreen', 'boolean', NULL, 0, NULL, NULL, 'Touchscreen capability', datetime('now'), datetime('now')),
-- Monitor attributes
(7, 2, 'Screen Size', 'number', 'inches', 1, 21, 34, 'Display diagonal size', datetime('now'), datetime('now')),
(8, 2, 'Resolution Width', 'number', 'px', 1, 1920, 7680, 'Horizontal resolution', datetime('now'), datetime('now')),
(9, 2, 'Resolution Height', 'number', 'px', 1, 1080, 4320, 'Vertical resolution', datetime('now'), datetime('now')),
(10, 2, 'Refresh Rate', 'number', 'Hz', 0, 60, 240, 'Screen refresh rate', datetime('now'), datetime('now')),
(11, 2, 'Panel Type', 'text', NULL, 0, NULL, NULL, 'IPS, VA, TN, OLED', datetime('now'), datetime('now')),
-- Smartphone attributes
(12, 3, 'RAM', 'number', 'GB', 1, 4, 24, 'Memory capacity', datetime('now'), datetime('now')),
(13, 3, 'Storage', 'number', 'GB', 1, 64, 1024, 'Internal storage', datetime('now'), datetime('now')),
(14, 3, 'Screen Size', 'number', 'inches', 1, 5, 7, 'Display diagonal size', datetime('now'), datetime('now')),
(15, 3, 'Battery', 'number', 'mAh', 0, 3000, 6000, 'Battery capacity', datetime('now'), datetime('now')),
(16, 3, 'Has 5G', 'boolean', NULL, 0, NULL, NULL, '5G connectivity', datetime('now'), datetime('now')),
-- Mouse attributes
(17, 4, 'DPI', 'number', 'DPI', 0, 800, 25600, 'Sensor resolution', datetime('now'), datetime('now')),
(18, 4, 'Is Wireless', 'boolean', NULL, 1, NULL, NULL, 'Wireless connectivity', datetime('now'), datetime('now')),
(19, 4, 'Battery Life', 'number', 'days', 0, 7, 365, 'Battery life in days', datetime('now'), datetime('now')),
-- Keyboard attributes
(20, 5, 'Switch Type', 'text', NULL, 1, NULL, NULL, 'Mechanical switch type', datetime('now'), datetime('now')),
(21, 5, 'Is Wireless', 'boolean', NULL, 0, NULL, NULL, 'Wireless connectivity', datetime('now'), datetime('now')),
(22, 5, 'Has RGB', 'boolean', NULL, 0, NULL, NULL, 'RGB backlight', datetime('now'), datetime('now'));

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
-- Laptops (for comparison scenarios)
(1, 'MacBook Pro 15"', 'APPLE-MBP15-2024', 'High-performance laptop with M3 Pro chip, 15-inch Liquid Retina display', 1, 1, 'each', 1, 7, 1, 'admin', datetime('now'), datetime('now')),
(2, 'Dell XPS 15', 'DELL-XPS15-9530', 'Premium Windows laptop with Intel Core i7, 15.6-inch OLED display', 2, 1, 'each', 1, 5, 1, 'admin', datetime('now'), datetime('now')),
(3, 'ThinkPad X1 Carbon', 'LENOVO-X1C-G11', 'Business ultrabook with Intel Core i5, 14-inch display', 6, 1, 'each', 5, 10, 1, 'admin', datetime('now'), datetime('now')),
(10, 'HP EliteBook 850', 'HP-EB850-G10', 'Business laptop with Intel vPro, 15-inch display - INCOMPLETE SPECS', 5, 1, 'each', 10, 14, 1, 'admin', datetime('now'), datetime('now')),
(11, 'Dell Latitude 5530', 'DELL-LAT5530', 'Mid-range business laptop with good value', 2, 1, 'each', 5, 7, 1, 'admin', datetime('now'), datetime('now')),
-- Monitors
(4, 'Dell UltraSharp U2720Q', 'DELL-U2720Q', 'Professional 27-inch 4K IPS monitor with USB-C hub', 2, 2, 'each', 2, 7, 1, 'admin', datetime('now'), datetime('now')),
(5, 'LG 27UK850-W', 'LG-27UK850', '27-inch 4K UHD IPS monitor with HDR10 support', 7, 2, 'each', 1, 5, 1, 'admin', datetime('now'), datetime('now')),
(12, 'Samsung M7', 'SAMS-M7-32', '32-inch 4K smart monitor with streaming apps - MISSING REFRESH RATE', 3, 2, 'each', 1, 5, 1, 'admin', datetime('now'), datetime('now')),
-- Smartphones
(6, 'Samsung Galaxy S24 Ultra', 'SAMS-S24U-256', 'Flagship Android smartphone with 256GB storage, S Pen', 3, 3, 'each', 1, 3, 1, 'admin', datetime('now'), datetime('now')),
(7, 'iPhone 15 Pro Max', 'APPLE-IP15PM-256', 'Premium iPhone with A17 Pro chip, 256GB storage, titanium design', 1, 3, 'each', 1, 2, 1, 'admin', datetime('now'), datetime('now')),
-- Peripherals
(8, 'Logitech MX Master 3S', 'LOGI-MXM3S-BLK', 'Ergonomic wireless mouse with 8K DPI sensor and quiet clicks', 4, 4, 'each', 10, 3, 1, 'admin', datetime('now'), datetime('now')),
(9, 'Logitech MX Mechanical', 'LOGI-MXMECH-BRN', 'Wireless mechanical keyboard with tactile brown switches and backlight', 4, 5, 'each', 5, 3, 1, 'admin', datetime('now'), datetime('now'));

-- Product Attributes (actual feature values for each product)
INSERT INTO product_attributes (id, product_id, specification_attribute_id, value_text, value_number, value_boolean, created_at, updated_at) VALUES
-- MacBook Pro 15" (Product 1) - Laptop
(1, 1, 1, NULL, 36, NULL, datetime('now'), datetime('now')),    -- RAM: 36 GB
(2, 1, 2, NULL, 512, NULL, datetime('now'), datetime('now')),   -- Storage: 512 GB
(3, 1, 3, NULL, 15.3, NULL, datetime('now'), datetime('now')),  -- Screen Size: 15.3 inches
(4, 1, 4, NULL, 12, NULL, datetime('now'), datetime('now')),    -- CPU Cores: 12
(5, 1, 5, 'SSD', NULL, NULL, datetime('now'), datetime('now')), -- Storage Type: SSD
(6, 1, 6, NULL, NULL, 0, datetime('now'), datetime('now')),     -- Has Touchscreen: false
-- Dell XPS 15 (Product 2) - Laptop
(7, 2, 1, NULL, 32, NULL, datetime('now'), datetime('now')),    -- RAM: 32 GB
(8, 2, 2, NULL, 1024, NULL, datetime('now'), datetime('now')),  -- Storage: 1024 GB
(9, 2, 3, NULL, 15.6, NULL, datetime('now'), datetime('now')),  -- Screen Size: 15.6 inches
(10, 2, 4, NULL, 8, NULL, datetime('now'), datetime('now')),    -- CPU Cores: 8
(11, 2, 5, 'SSD', NULL, NULL, datetime('now'), datetime('now')),-- Storage Type: SSD
(12, 2, 6, NULL, NULL, 1, datetime('now'), datetime('now')),    -- Has Touchscreen: true
-- ThinkPad X1 Carbon (Product 3) - Laptop
(13, 3, 1, NULL, 16, NULL, datetime('now'), datetime('now')),   -- RAM: 16 GB
(14, 3, 2, NULL, 512, NULL, datetime('now'), datetime('now')),  -- Storage: 512 GB
(15, 3, 3, NULL, 14, NULL, datetime('now'), datetime('now')),   -- Screen Size: 14 inches
(16, 3, 4, NULL, 10, NULL, datetime('now'), datetime('now')),   -- CPU Cores: 10
(17, 3, 5, 'SSD', NULL, NULL, datetime('now'), datetime('now')),-- Storage Type: SSD
(18, 3, 6, NULL, NULL, 0, datetime('now'), datetime('now')),    -- Has Touchscreen: false
-- Dell UltraSharp U2720Q (Product 4) - Monitor
(19, 4, 7, NULL, 27, NULL, datetime('now'), datetime('now')),   -- Screen Size: 27 inches
(20, 4, 8, NULL, 3840, NULL, datetime('now'), datetime('now')), -- Resolution Width: 3840 px
(21, 4, 9, NULL, 2160, NULL, datetime('now'), datetime('now')), -- Resolution Height: 2160 px
(22, 4, 10, NULL, 60, NULL, datetime('now'), datetime('now')),  -- Refresh Rate: 60 Hz
(23, 4, 11, 'IPS', NULL, NULL, datetime('now'), datetime('now')),-- Panel Type: IPS
-- LG 27UK850-W (Product 5) - Monitor
(24, 5, 7, NULL, 27, NULL, datetime('now'), datetime('now')),   -- Screen Size: 27 inches
(25, 5, 8, NULL, 3840, NULL, datetime('now'), datetime('now')), -- Resolution Width: 3840 px
(26, 5, 9, NULL, 2160, NULL, datetime('now'), datetime('now')), -- Resolution Height: 2160 px
(27, 5, 10, NULL, 60, NULL, datetime('now'), datetime('now')),  -- Refresh Rate: 60 Hz
(28, 5, 11, 'IPS', NULL, NULL, datetime('now'), datetime('now')),-- Panel Type: IPS
-- Samsung Galaxy S24 Ultra (Product 6) - Smartphone
(29, 6, 12, NULL, 12, NULL, datetime('now'), datetime('now')),  -- RAM: 12 GB
(30, 6, 13, NULL, 256, NULL, datetime('now'), datetime('now')), -- Storage: 256 GB
(31, 6, 14, NULL, 6.8, NULL, datetime('now'), datetime('now')), -- Screen Size: 6.8 inches
(32, 6, 15, NULL, 5000, NULL, datetime('now'), datetime('now')),-- Battery: 5000 mAh
(33, 6, 16, NULL, NULL, 1, datetime('now'), datetime('now')),   -- Has 5G: true
-- iPhone 15 Pro Max (Product 7) - Smartphone
(34, 7, 12, NULL, 8, NULL, datetime('now'), datetime('now')),   -- RAM: 8 GB
(35, 7, 13, NULL, 256, NULL, datetime('now'), datetime('now')), -- Storage: 256 GB
(36, 7, 14, NULL, 6.7, NULL, datetime('now'), datetime('now')), -- Screen Size: 6.7 inches
(37, 7, 15, NULL, 4441, NULL, datetime('now'), datetime('now')),-- Battery: 4441 mAh
(38, 7, 16, NULL, NULL, 1, datetime('now'), datetime('now')),   -- Has 5G: true
-- Logitech MX Master 3S (Product 8) - Mouse
(39, 8, 17, NULL, 8000, NULL, datetime('now'), datetime('now')),-- DPI: 8000
(40, 8, 18, NULL, NULL, 1, datetime('now'), datetime('now')),   -- Is Wireless: true
(41, 8, 19, NULL, 70, NULL, datetime('now'), datetime('now')),  -- Battery Life: 70 days
-- Logitech MX Mechanical (Product 9) - Keyboard
(42, 9, 20, 'Tactile Brown', NULL, NULL, datetime('now'), datetime('now')),-- Switch Type: Tactile Brown
(43, 9, 21, NULL, NULL, 1, datetime('now'), datetime('now')),   -- Is Wireless: true
(44, 9, 22, NULL, NULL, 1, datetime('now'), datetime('now')),   -- Has RGB: true
-- HP EliteBook 850 (Product 10) - Laptop - INCOMPLETE (missing required Storage attribute)
(45, 10, 1, NULL, 16, NULL, datetime('now'), datetime('now')),  -- RAM: 16 GB
-- Missing: Storage (REQUIRED attribute not set)
(46, 10, 3, NULL, 15, NULL, datetime('now'), datetime('now')),  -- Screen Size: 15 inches
(47, 10, 4, NULL, 8, NULL, datetime('now'), datetime('now')),   -- CPU Cores: 8
(48, 10, 5, 'SSD', NULL, NULL, datetime('now'), datetime('now')),-- Storage Type: SSD (but no storage size!)
(49, 10, 6, NULL, NULL, 0, datetime('now'), datetime('now')),   -- Has Touchscreen: false
-- Dell Latitude 5530 (Product 11) - Laptop - COMPLETE (budget option)
(50, 11, 1, NULL, 16, NULL, datetime('now'), datetime('now')),  -- RAM: 16 GB
(51, 11, 2, NULL, 256, NULL, datetime('now'), datetime('now')), -- Storage: 256 GB
(52, 11, 3, NULL, 15, NULL, datetime('now'), datetime('now')),  -- Screen Size: 15 inches
(53, 11, 4, NULL, 6, NULL, datetime('now'), datetime('now')),   -- CPU Cores: 6
(54, 11, 5, 'SSD', NULL, NULL, datetime('now'), datetime('now')),-- Storage Type: SSD
-- Samsung M7 (Product 12) - Monitor - INCOMPLETE (missing required attributes)
(55, 12, 7, NULL, 32, NULL, datetime('now'), datetime('now')),  -- Screen Size: 32 inches
(56, 12, 8, NULL, 3840, NULL, datetime('now'), datetime('now')), -- Resolution Width: 3840 px
(57, 12, 9, NULL, 2160, NULL, datetime('now'), datetime('now')), -- Resolution Height: 2160 px
-- Missing: Refresh Rate is NOT SET (optional, but good for comparison)
(58, 12, 11, 'VA', NULL, NULL, datetime('now'), datetime('now'));-- Panel Type: VA

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
-- LAPTOP QUOTES - For comparison matrix demonstration
-- Product 1: MacBook Pro 15" (COMPLETE attributes, HIGH price)
(1, 1, 1, 2, 2499.00, 'USD', 2499.00, 1.0, 1, 'active', datetime('now'), 'MacBook Pro with educational discount - updated price', 'sales_team', datetime('now'), datetime('now')),
(2, 2, 1, 1, 2399.00, 'USD', 2399.00, 1.0, 3, 'active', datetime('now'), 'Amazon Business bulk pricing - minimum 3 units', 'procurement', datetime('now'), datetime('now')),
-- Product 2: Dell XPS 15 (COMPLETE attributes, MEDIUM-HIGH price)
(3, 1, 2, 1, 1599.99, 'USD', 1599.99, 1.0, 1, 'superseded', datetime('now', '-7 days'), 'Dell XPS on sale - OLD QUOTE', 'sales_team', datetime('now', '-7 days'), datetime('now')),
(4, 4, 2, 2, 1699.00, 'USD', 1699.00, 1.0, 5, 'active', datetime('now'), 'CDW corporate pricing - minimum 5 units for discount', 'procurement', datetime('now'), datetime('now')),
(13, 2, 2, 1, 1649.00, 'USD', 1649.00, 1.0, 1, 'active', datetime('now'), 'Amazon single unit price - Prime eligible', 'procurement', datetime('now'), datetime('now')),
-- Product 3: ThinkPad X1 Carbon (COMPLETE attributes, MEDIUM price, 14-inch)
(14, 4, 3, 1, 1299.00, 'USD', 1299.00, 1.0, 5, 'active', datetime('now'), 'CDW corporate discount - 5+ units', 'procurement', datetime('now'), datetime('now')),
(15, 1, 3, 1, 1399.00, 'USD', 1399.00, 1.0, 1, 'active', datetime('now'), 'Best Buy business pricing', 'sales_team', datetime('now'), datetime('now')),
-- Product 10: HP EliteBook 850 (INCOMPLETE - missing Storage, LOW price to tempt buyers)
(16, 4, 10, 1, 1099.00, 'USD', 1099.00, 1.0, 10, 'active', datetime('now'), 'HP EliteBook bulk pricing - INCOMPLETE SPECS WARNING', 'procurement', datetime('now'), datetime('now')),
(17, 2, 10, 1, 1149.00, 'USD', 1149.00, 1.0, 5, 'active', datetime('now'), 'Amazon Business - verify specs before ordering', 'procurement', datetime('now'), datetime('now')),
-- Product 11: Dell Latitude 5530 (COMPLETE attributes, LOWEST price - budget option)
(18, 4, 11, 1, 899.00, 'USD', 899.00, 1.0, 5, 'active', datetime('now'), 'Dell Latitude budget option - complete specs', 'procurement', datetime('now'), datetime('now')),
(19, 2, 11, 1, 949.00, 'USD', 949.00, 1.0, 1, 'active', datetime('now'), 'Amazon single unit - good value for money', 'procurement', datetime('now'), datetime('now')),
(20, 1, 11, 1, 979.00, 'USD', 979.00, 1.0, 1, 'active', datetime('now'), 'Best Buy retail price', 'sales_team', datetime('now'), datetime('now')),
-- MONITOR QUOTES - For comparison matrix demonstration
(5, 3, 4, 1, 649.99, 'USD', 649.99, 1.0, 2, 'accepted', datetime('now', '-5 days'), 'Monitor with free shipping - ACCEPTED', 'manager', datetime('now', '-5 days'), datetime('now', '-5 days')),
(6, 2, 5, 1, 599.00, 'USD', 599.00, 1.0, 1, 'active', datetime('now'), 'Amazon Prime discount', 'procurement', datetime('now'), datetime('now')),
(21, 1, 5, 1, 629.00, 'USD', 629.00, 1.0, 1, 'active', datetime('now'), 'Best Buy retail - price match available', 'sales_team', datetime('now'), datetime('now')),
(22, 2, 12, 1, 499.00, 'USD', 499.00, 1.0, 1, 'active', datetime('now'), 'Samsung M7 Smart Monitor - MISSING refresh rate spec', 'procurement', datetime('now'), datetime('now')),
(23, 1, 12, 1, 529.00, 'USD', 529.00, 1.0, 1, 'active', datetime('now'), 'Samsung M7 at Best Buy - verify specs', 'sales_team', datetime('now'), datetime('now')),
-- SMARTPHONE QUOTES
(7, 1, 7, 1, 1199.00, 'USD', 1199.00, 1.0, 1, 'active', datetime('now'), 'iPhone with trade-in credit', 'sales_team', datetime('now'), datetime('now')),
(8, 5, 6, 1, 5999.00, 'CNY', 839.86, 0.14, 10, 'active', datetime('now'), 'Bulk order available - minimum 10 units', 'procurement', datetime('now'), datetime('now')),
-- PERIPHERAL QUOTES
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

-- Vendor Ratings (performance tracking)
INSERT INTO vendor_ratings (id, vendor_id, purchase_order_id, price_rating, quality_rating, delivery_rating, service_rating, comments, rated_by, created_at, updated_at) VALUES
-- Ratings for completed purchase orders
(1, 2, 2, 4, 5, 5, 5, 'Excellent service from B&H Photo. Camera arrived well-packaged and on time. Competitive pricing.', 'procurement@company.com', datetime('now', '-2 days'), datetime('now', '-2 days')),
(2, 3, 3, 5, 5, 4, 5, 'CDW provided great bulk pricing on keyboards. Minor delay in shipping but overall excellent transaction.', 'admin@company.com', datetime('now', '-4 days'), datetime('now', '-4 days')),
(3, 4, 4, 3, 4, 5, 4, 'Newegg had good prices on monitors. Quality is solid, delivered ahead of schedule.', 'procurement@company.com', datetime('now', '-7 days'), datetime('now', '-7 days')),
(4, 5, 5, 2, 3, 2, 3, 'Alibaba pricing was competitive but shipping took longer than expected. Product quality acceptable but not premium.', 'sourcing@company.com', datetime('now', '-10 days'), datetime('now', '-10 days')),
-- General vendor ratings (not tied to specific POs)
(5, 1, NULL, 5, 5, 5, 5, 'Best Buy is our go-to for quick procurement needs. Consistently reliable.', 'admin@company.com', datetime('now', '-15 days'), datetime('now', '-15 days')),
(6, 2, NULL, 5, 5, 5, 5, 'B&H Photo has been excellent for all our camera and video equipment needs.', 'media@company.com', datetime('now', '-20 days'), datetime('now', '-20 days')),
(7, 3, NULL, 4, 5, 4, 5, 'CDW is very professional and offers good enterprise support.', 'it@company.com', datetime('now', '-25 days'), datetime('now', '-25 days'));
