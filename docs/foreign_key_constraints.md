# Foreign Key Constraints Implementation

## Summary

Added database-level foreign key constraints to prevent orphaned records and ensure referential integrity across all relationships in the buyer application.

## Changes Made

### 1. Model Updates (`internal/models/models.go`)

Added GORM constraint tags to all foreign key relationships:

#### RESTRICT Constraints (Prevent deletion of referenced records)
- `Vendor.Quotes` → Prevents deleting vendors that have quotes
- `Brand.Products` → Prevents deleting brands that have products
- `Quote.Vendor` → Prevents deleting vendors referenced by quotes
- `Quote.Product` → Prevents deleting products referenced by quotes
- `RequisitionItem.Specification` → Prevents deleting specifications referenced by requisition items

#### CASCADE Constraints (Auto-delete dependent records)
- `Product.Quotes` → Deletes all quotes when product is deleted
- `Requisition.Items` → Deletes all requisition items when requisition is deleted
- `RequisitionItem.Requisition` → Deletes items when parent requisition is deleted

#### SET NULL Constraints (Nullify foreign keys)
- `Specification.Products` → Sets `Product.SpecificationID` to NULL when specification is deleted
- `Product.Specification` → Allows products to exist without a specification

### 2. Schema Change

Changed `Product.SpecificationID` from `uint` to `*uint` (nullable) to support SET NULL behavior:
- Products can now exist without a specification
- When a specification is deleted, products retain their brand but lose the specification reference

### 3. Database Configuration (`internal/config/config.go`)

Enabled foreign key constraint enforcement for SQLite:
```go
db.Exec("PRAGMA foreign_keys = ON")
```

This is critical for SQLite as foreign key constraints are **disabled by default**.

### 4. Service Layer Updates (`internal/services/product.go`)

Updated `ProductService` to handle nullable `SpecificationID`:
- Simplified assignment logic to use pointer directly
- Removed manual zero-value checks (now handled by NULL)

### 5. Comprehensive Testing (`internal/models/constraints_test.go`)

Added 5 test cases verifying constraint behavior:
1. **RESTRICT on Brand deletion** - Prevents deleting brands with products
2. **RESTRICT on Vendor deletion** - Prevents deleting vendors with quotes
3. **CASCADE on Requisition deletion** - Auto-deletes requisition items
4. **CASCADE on Product deletion** - Auto-deletes quotes
5. **SET NULL on Specification deletion** - Nullifies product specification references

All tests pass [x]

## Constraint Strategy

| Relationship | Constraint | Rationale |
|-------------|-----------|-----------|
| Brand → Product | RESTRICT | Products cannot exist without a brand |
| Vendor → Quote | RESTRICT | Preserve quote history even if vendor changes |
| Product → Quote | CASCADE | Quotes are meaningless without the product |
| Specification → Product | SET NULL | Products can exist without categorization |
| Requisition → Items | CASCADE | Line items are part of requisition lifecycle |

## Migration Notes

### Existing Databases

Users with existing databases will need to:

1. **Backup current database**:
   ```bash
   cp ~/.buyer/buyer.db ~/.buyer/buyer.db.backup
   ```

2. **Restart application** - GORM AutoMigrate will add constraints to new tables

3. **For existing data** (optional cleanup):
   ```sql
   -- Find orphaned products (brand was deleted)
   SELECT * FROM products WHERE brand_id NOT IN (SELECT id FROM brands);

   -- Find orphaned quotes (product or vendor was deleted)
   SELECT * FROM quotes WHERE product_id NOT IN (SELECT id FROM products)
                          OR vendor_id NOT IN (SELECT id FROM vendors);
   ```

### Fresh Installations

All constraints are automatically created with the schema - no manual intervention needed.

## Breaking Changes

[!] **API Change**: `Product.SpecificationID` is now `*uint` instead of `uint`

Code interacting with products must handle the pointer:
```go
// Before
if product.SpecificationID == 0 { ... }

// After
if product.SpecificationID == nil { ... }
// or
if product.SpecificationID != nil && *product.SpecificationID == 0 { ... }
```

The `ProductService` API already used `*uint`, so no service-level changes required.

## Benefits

1. **Data Integrity** - Impossible to create orphaned records at database level
2. **Cascade Cleanup** - Automatically removes dependent data when parents are deleted
3. **Error Prevention** - Database enforces constraints before application logic runs
4. **Audit Trail** - RESTRICT constraints prevent accidental deletion of referenced entities
5. **Flexibility** - Products can exist without specifications (SET NULL)

## Testing

Run constraint tests:
```bash
go test -v ./internal/models -run TestForeignKeyConstraints
```

Run all tests:
```bash
make test
```

All tests pass with the new constraints [x]
