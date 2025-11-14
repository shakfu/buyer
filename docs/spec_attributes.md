# Specification Attributes Guide

This guide explains how to use and query the `SpecificationAttribute` and `ProductAttribute` models to define comparable features for product specifications and store actual values for each product.

## Table of Contents

1. [Overview](#overview)
2. [Data Model](#data-model)
3. [Creating Specification Attributes](#creating-specification-attributes)
4. [Adding Product Attributes](#adding-product-attributes)
5. [Querying Patterns](#querying-patterns)
6. [Use Cases](#use-cases)
7. [Best Practices](#best-practices)

## Overview

The attribute system allows you to:
- Define standardized features for each specification type (e.g., "Laptop" has RAM, Storage, Screen Size)
- Store actual feature values for each product (e.g., MacBook Pro has 36GB RAM, 512GB Storage)
- Compare products by their features using powerful SQL queries
- Validate data integrity with type checking and constraints

**Architecture:**
```
Specification (e.g., "Laptop - 15 inch")
    └── SpecificationAttribute (e.g., "RAM", type: number, unit: GB)
            └── ProductAttribute (e.g., Product: MacBook Pro, value: 36)
```

## Data Model

### SpecificationAttribute

Defines what attributes a specification type should have.

**Fields:**
- `specification_id` (uint) - FK to Specification
- `name` (string) - Attribute name (e.g., "RAM", "Screen Size")
- `data_type` (string) - One of: `text`, `number`, `boolean`
- `unit` (string, optional) - Unit of measurement (e.g., "GB", "inches", "Hz")
- `is_required` (bool) - Whether products must have this attribute
- `min_value` (float64, optional) - Minimum value for number types
- `max_value` (float64, optional) - Maximum value for number types
- `description` (string) - Help text explaining the attribute

**Example:**
```go
models.SpecificationAttribute{
    SpecificationID: 1,           // Laptop specification
    Name:            "RAM",
    DataType:        "number",
    Unit:            "GB",
    IsRequired:      true,
    MinValue:        ptr(4.0),
    MaxValue:        ptr(128.0),
    Description:     "Memory capacity",
}
```

### ProductAttribute

Stores actual attribute values for a specific product.

**Fields:**
- `product_id` (uint) - FK to Product
- `specification_attribute_id` (uint) - FK to SpecificationAttribute
- `value_text` (*string) - For text values
- `value_number` (*float64) - For numeric values
- `value_boolean` (*bool) - For boolean values

**Constraints:**
- Exactly ONE value field must be set (text, number, or boolean)
- The value type must match the SpecificationAttribute's `data_type`
- Numeric values must be within min/max constraints if defined

**Example:**
```go
models.ProductAttribute{
    ProductID:                1,    // MacBook Pro
    SpecificationAttributeID: 1,    // RAM attribute
    ValueNumber:              ptr(36.0), // 36 GB
}
```

## Creating Specification Attributes

### Using Raw SQL

```sql
-- Define attributes for "Laptop" specification (spec_id = 1)
INSERT INTO specification_attributes (
    specification_id, name, data_type, unit,
    is_required, min_value, max_value, description,
    created_at, updated_at
) VALUES
    (1, 'RAM', 'number', 'GB', 1, 4, 128, 'Memory capacity', datetime('now'), datetime('now')),
    (1, 'Storage', 'number', 'GB', 1, 128, 4096, 'Storage capacity', datetime('now'), datetime('now')),
    (1, 'Storage Type', 'text', NULL, 1, NULL, NULL, 'SSD or HDD', datetime('now'), datetime('now')),
    (1, 'Has Touchscreen', 'boolean', NULL, 0, NULL, NULL, 'Touchscreen capability', datetime('now'), datetime('now'));
```

### Using GORM

```go
db.Create(&models.SpecificationAttribute{
    SpecificationID: laptopSpecID,
    Name:            "RAM",
    DataType:        "number",
    Unit:            "GB",
    IsRequired:      true,
    MinValue:        ptr(4.0),
    MaxValue:        ptr(128.0),
    Description:     "Memory capacity",
})
```

### Data Types

**1. Number** - For numeric values (RAM, storage, screen size, etc.)
```go
{
    Name:      "RAM",
    DataType:  "number",
    Unit:      "GB",
    MinValue:  ptr(4.0),
    MaxValue:  ptr(128.0),
}
```

**2. Text** - For string values (storage type, panel type, etc.)
```go
{
    Name:        "Storage Type",
    DataType:    "text",
    Description: "SSD, HDD, or Hybrid",
}
```

**3. Boolean** - For yes/no flags (has touchscreen, wireless, etc.)
```go
{
    Name:        "Has Touchscreen",
    DataType:    "boolean",
    Description: "Touchscreen capability",
}
```

## Adding Product Attributes

### Using Raw SQL

```sql
-- Add attributes for MacBook Pro (product_id = 1)
INSERT INTO product_attributes (
    product_id, specification_attribute_id,
    value_text, value_number, value_boolean,
    created_at, updated_at
) VALUES
    (1, 1, NULL, 36, NULL, datetime('now'), datetime('now')),     -- RAM: 36 GB
    (1, 2, NULL, 512, NULL, datetime('now'), datetime('now')),    -- Storage: 512 GB
    (1, 5, 'SSD', NULL, NULL, datetime('now'), datetime('now')),  -- Storage Type: SSD
    (1, 6, NULL, NULL, 0, datetime('now'), datetime('now'));      -- Has Touchscreen: false
```

### Using GORM

```go
// Number value
db.Create(&models.ProductAttribute{
    ProductID:                macbookProID,
    SpecificationAttributeID: ramAttrID,
    ValueNumber:              ptr(36.0),
})

// Text value
db.Create(&models.ProductAttribute{
    ProductID:                macbookProID,
    SpecificationAttributeID: storageTypeAttrID,
    ValueText:                ptr("SSD"),
})

// Boolean value
db.Create(&models.ProductAttribute{
    ProductID:                macbookProID,
    SpecificationAttributeID: touchscreenAttrID,
    ValueBoolean:             ptr(false),
})
```

## Querying Patterns

### 1. Get All Attributes for a Product

```sql
SELECT
    p.name as product_name,
    sa.name as attribute_name,
    COALESCE(
        pa.value_text,
        CAST(pa.value_number AS TEXT),
        CAST(pa.value_boolean AS TEXT)
    ) as value,
    sa.unit
FROM products p
JOIN product_attributes pa ON p.id = pa.product_id
JOIN specification_attributes sa ON pa.specification_attribute_id = sa.id
WHERE p.id = 1  -- MacBook Pro
ORDER BY sa.name;
```

**Result:**
```
product_name     attribute_name   value  unit
---------------  ---------------  -----  ------
MacBook Pro 15"  CPU Cores        12.0   cores
MacBook Pro 15"  Has Touchscreen  0
MacBook Pro 15"  RAM              36.0   GB
MacBook Pro 15"  Screen Size      15.3   inches
MacBook Pro 15"  Storage          512.0  GB
MacBook Pro 15"  Storage Type     SSD
```

### 2. Compare Products by Specific Attributes

```sql
-- Compare laptops by RAM, Storage, and Screen Size
SELECT
    p.name,
    ram.value_number as ram_gb,
    storage.value_number as storage_gb,
    screen.value_number as screen_inches
FROM products p
JOIN specification_attributes ram_attr
    ON ram_attr.specification_id = p.specification_id
    AND ram_attr.name = 'RAM'
JOIN product_attributes ram
    ON ram.product_id = p.id
    AND ram.specification_attribute_id = ram_attr.id
JOIN specification_attributes storage_attr
    ON storage_attr.specification_id = p.specification_id
    AND storage_attr.name = 'Storage'
JOIN product_attributes storage
    ON storage.product_id = p.id
    AND storage.specification_attribute_id = storage_attr.id
JOIN specification_attributes screen_attr
    ON screen_attr.specification_id = p.specification_id
    AND screen_attr.name = 'Screen Size'
JOIN product_attributes screen
    ON screen.product_id = p.id
    AND screen.specification_attribute_id = screen_attr.id
WHERE p.specification_id = 1  -- Laptops only
ORDER BY ram.value_number DESC;
```

**Result:**
```
name                ram_gb  storage_gb  screen_inches
------------------  ------  ----------  -------------
MacBook Pro 15"     36.0    512.0       15.3
Dell XPS 15         32.0    1024.0      15.6
ThinkPad X1 Carbon  16.0    512.0       14.0
```

### 3. Filter Products by Attribute Values

```sql
-- Find all laptops with 32GB+ RAM and SSD storage
SELECT
    p.name,
    ram.value_number as ram_gb,
    storage_type.value_text as storage
FROM products p
JOIN product_attributes ram ON ram.product_id = p.id
JOIN specification_attributes ram_attr
    ON ram_attr.id = ram.specification_attribute_id
    AND ram_attr.name = 'RAM'
JOIN product_attributes storage_type ON storage_type.product_id = p.id
JOIN specification_attributes storage_attr
    ON storage_attr.id = storage_type.specification_attribute_id
    AND storage_attr.name = 'Storage Type'
WHERE p.specification_id = 1
    AND ram.value_number >= 32
    AND storage_type.value_text = 'SSD';
```

### 4. Get All Attributes Defined for a Specification

```sql
-- List all attributes for "Laptop" specification
SELECT
    sa.name,
    sa.data_type,
    sa.unit,
    sa.is_required,
    sa.min_value,
    sa.max_value,
    sa.description
FROM specification_attributes sa
JOIN specifications s ON s.id = sa.specification_id
WHERE s.name LIKE 'Laptop%'
ORDER BY sa.is_required DESC, sa.name;
```

### 5. Find Products Missing Required Attributes

```sql
-- Find products that don't have all required attributes
SELECT
    p.id,
    p.name,
    s.name as specification,
    sa.name as missing_attribute
FROM products p
JOIN specifications s ON s.id = p.specification_id
JOIN specification_attributes sa
    ON sa.specification_id = s.id
    AND sa.is_required = 1
LEFT JOIN product_attributes pa
    ON pa.product_id = p.id
    AND pa.specification_attribute_id = sa.id
WHERE pa.id IS NULL;
```

### 6. Compare Products with Quotes

```sql
-- Compare laptop specs alongside best quotes
SELECT
    p.name,
    b.name as brand,
    ram.value_number as ram_gb,
    storage.value_number as storage_gb,
    MIN(q.converted_price) as best_price_usd
FROM products p
JOIN brands b ON b.id = p.brand_id
LEFT JOIN quotes q ON q.product_id = p.id AND q.status = 'active'
JOIN product_attributes ram ON ram.product_id = p.id
JOIN specification_attributes ram_attr
    ON ram_attr.id = ram.specification_attribute_id
    AND ram_attr.name = 'RAM'
JOIN product_attributes storage ON storage.product_id = p.id
JOIN specification_attributes storage_attr
    ON storage_attr.id = storage.specification_attribute_id
    AND storage_attr.name = 'Storage'
WHERE p.specification_id = 1  -- Laptops
GROUP BY p.id, p.name, b.name, ram.value_number, storage.value_number
ORDER BY ram.value_number DESC, best_price_usd ASC;
```

### 7. Aggregate Statistics

```sql
-- Get average specs for a specification type
SELECT
    sa.name as attribute,
    ROUND(AVG(pa.value_number), 2) as avg_value,
    MIN(pa.value_number) as min_value,
    MAX(pa.value_number) as max_value,
    sa.unit
FROM specification_attributes sa
JOIN product_attributes pa ON pa.specification_attribute_id = sa.id
JOIN products p ON p.id = pa.product_id
WHERE sa.specification_id = 1  -- Laptops
    AND sa.data_type = 'number'
GROUP BY sa.id, sa.name, sa.unit
ORDER BY sa.name;
```

## Use Cases

### 1. Product Comparison Dashboard

Show side-by-side comparison of products:

```sql
-- Pivot-style comparison for specific products
SELECT
    sa.name as feature,
    MAX(CASE WHEN p.id = 1 THEN
        COALESCE(pa.value_text,
                 CAST(pa.value_number AS TEXT),
                 CAST(pa.value_boolean AS TEXT))
    END) as macbook_pro,
    MAX(CASE WHEN p.id = 2 THEN
        COALESCE(pa.value_text,
                 CAST(pa.value_number AS TEXT),
                 CAST(pa.value_boolean AS TEXT))
    END) as dell_xps,
    sa.unit
FROM specification_attributes sa
JOIN product_attributes pa ON pa.specification_attribute_id = sa.id
JOIN products p ON p.id = pa.product_id
WHERE p.id IN (1, 2)  -- MacBook Pro and Dell XPS
GROUP BY sa.id, sa.name, sa.unit
ORDER BY sa.name;
```

### 2. Requisition Matching

Find products that meet requisition requirements:

```sql
-- Find laptops with at least 32GB RAM for a requisition
SELECT
    p.id,
    p.name,
    ram.value_number as ram_gb,
    storage.value_number as storage_gb,
    MIN(q.converted_price) as best_price
FROM products p
JOIN product_attributes ram ON ram.product_id = p.id
JOIN specification_attributes ram_attr
    ON ram_attr.id = ram.specification_attribute_id
    AND ram_attr.name = 'RAM'
JOIN product_attributes storage ON storage.product_id = p.id
JOIN specification_attributes storage_attr
    ON storage_attr.id = storage.specification_attribute_id
    AND storage_attr.name = 'Storage'
LEFT JOIN quotes q ON q.product_id = p.id AND q.status = 'active'
WHERE p.specification_id = 1
    AND ram.value_number >= 32      -- Requirement: 32GB+ RAM
    AND storage.value_number >= 512  -- Requirement: 512GB+ Storage
GROUP BY p.id, p.name, ram.value_number, storage.value_number
HAVING best_price IS NOT NULL
ORDER BY best_price ASC;
```

### 3. Specification Template Creation

Create a new specification with standard attributes:

```sql
-- Copy attributes from one specification to another
INSERT INTO specification_attributes (
    specification_id, name, data_type, unit,
    is_required, min_value, max_value, description,
    created_at, updated_at
)
SELECT
    2 as specification_id,  -- New specification
    name, data_type, unit,
    is_required, min_value, max_value, description,
    datetime('now'), datetime('now')
FROM specification_attributes
WHERE specification_id = 1  -- Source specification (Laptop)
    AND name IN ('RAM', 'Storage');  -- Only copy these attributes
```

### 4. Data Validation Report

Check for attribute values outside valid ranges:

```sql
-- Find attribute values that violate min/max constraints
SELECT
    p.name as product,
    sa.name as attribute,
    pa.value_number as value,
    sa.min_value,
    sa.max_value,
    sa.unit
FROM product_attributes pa
JOIN specification_attributes sa ON sa.id = pa.specification_attribute_id
JOIN products p ON p.id = pa.product_id
WHERE sa.data_type = 'number'
    AND (
        (sa.min_value IS NOT NULL AND pa.value_number < sa.min_value) OR
        (sa.max_value IS NOT NULL AND pa.value_number > sa.max_value)
    );
```

## Best Practices

### 1. Naming Conventions

- Use consistent, human-readable names: "RAM", "Screen Size", "Storage Type"
- Avoid abbreviations unless standard: "DPI" is OK, "SCR_SZ" is not
- Use title case for attribute names
- Be specific: "Screen Size" not just "Size"

### 2. Data Type Selection

**Use `number` for:**
- Quantities: RAM, storage, screen size, battery capacity
- Measurements: weight, dimensions, DPI, refresh rate
- Countable items: CPU cores, USB ports

**Use `text` for:**
- Categories: storage type (SSD/HDD), panel type (IPS/VA)
- Model numbers or codes
- Descriptive values that don't need comparison

**Use `boolean` for:**
- Yes/no features: touchscreen, wireless, backlit keyboard
- Presence of features: "Has 5G", "Has USB-C"

### 3. Units

- Always specify units for numeric attributes
- Use standard abbreviations: GB, inches, Hz, mAh, cores
- Be consistent across related attributes
- Example: Use "GB" for both RAM and Storage, not "GB" and "gigabytes"

### 4. Validation Ranges

- Set realistic min/max values for numeric attributes
- Consider future-proofing: RAM might go higher than current max
- Use constraints to catch data entry errors
- Example: Screen size 5-17 inches for laptops, 21-34 inches for monitors

### 5. Required vs Optional

**Mark as required:**
- Core defining features: RAM, storage, screen size for laptops
- Features critical for comparison
- Attributes needed for procurement decisions

**Mark as optional:**
- Nice-to-have specs: touchscreen, specific CPU model
- Variable features: some products have it, some don't
- Emerging features: 5G, specific connectivity options

### 6. Query Performance

**Optimize common queries:**
```sql
-- Create indexes for frequently queried attributes
CREATE INDEX idx_product_attr_by_attr
    ON product_attributes(specification_attribute_id, product_id);
CREATE INDEX idx_spec_attr_by_name
    ON specification_attributes(specification_id, name);
```

**Use preloading in GORM:**
```go
// Load product with all attributes
db.Preload("Attributes.SpecificationAttribute").Find(&product, id)
```

### 7. Data Consistency

**Prevent orphaned attributes:**
- Foreign key constraints ensure referential integrity
- CASCADE delete from Specification to SpecificationAttribute
- RESTRICT delete from SpecificationAttribute to ProductAttribute

**Validate before insert:**
```go
// Check that product's specification has the attribute
var count int64
db.Model(&models.SpecificationAttribute{}).
    Where("id = ? AND specification_id = ?",
          attrID, product.SpecificationID).
    Count(&count)
if count == 0 {
    return errors.New("attribute does not belong to product's specification")
}
```

### 8. Helper Functions

**Create reusable query builders:**
```go
// GetProductAttributeValue returns the value for a specific attribute
func GetProductAttributeValue(db *gorm.DB, productID uint, attrName string) (interface{}, error) {
    var attr models.ProductAttribute
    err := db.
        Joins("JOIN specification_attributes sa ON sa.id = product_attributes.specification_attribute_id").
        Where("product_attributes.product_id = ? AND sa.name = ?", productID, attrName).
        Preload("SpecificationAttribute").
        First(&attr).Error

    if err != nil {
        return nil, err
    }

    // Return the appropriate value based on type
    if attr.ValueNumber != nil {
        return *attr.ValueNumber, nil
    }
    if attr.ValueText != nil {
        return *attr.ValueText, nil
    }
    if attr.ValueBoolean != nil {
        return *attr.ValueBoolean, nil
    }
    return nil, errors.New("no value set")
}
```

### 9. Documentation

- Document each attribute's purpose in the `description` field
- Keep attribute definitions consistent across similar specifications
- Maintain a reference guide for attribute meanings
- Document valid values for text attributes (e.g., "IPS, VA, TN, OLED")

### 10. Testing

**Test attribute constraints:**
```go
// Test that min/max validation works
attr := &models.ProductAttribute{
    ProductID:                productID,
    SpecificationAttributeID: ramAttrID,  // min: 4, max: 128
    ValueNumber:              ptr(256.0), // Above max
}
err := db.Create(attr).Error
// Should fail validation
```

**Test data type matching:**
```go
// Test that text attribute rejects number
attr := &models.ProductAttribute{
    ProductID:                productID,
    SpecificationAttributeID: storageTypeAttrID,  // data_type: text
    ValueNumber:              ptr(500.0),         // Wrong type
}
err := db.Create(attr).Error
// Should fail validation
```

## Common Patterns Summary

| Use Case | Join Pattern | Key Consideration |
|----------|--------------|-------------------|
| Get all attributes for a product | Product → ProductAttribute → SpecificationAttribute | Use COALESCE for unified value field |
| Compare products | Self-join ProductAttribute with different product_id | Join on attribute name for alignment |
| Filter by attributes | Product → ProductAttribute → filter on value | Index on specification_attribute_id |
| Find missing attributes | LEFT JOIN with NULL check | Check is_required flag |
| Aggregate specs | GROUP BY specification_attribute_id | Only for numeric attributes |
| Attribute + price comparison | Product → ProductAttribute + Quotes | Use MIN for best price |

## Future Enhancements

Potential additions to consider:

1. **Attribute Groups**: Group related attributes (e.g., "Display", "Performance")
2. **Enumerated Values**: Define valid values for text attributes
3. **Computed Attributes**: Derive values from other attributes
4. **Versioning**: Track attribute value changes over time
5. **Attribute Templates**: Predefined sets for common specification types
6. **Weighted Scoring**: Assign importance weights for comparison algorithms
