# Quote Comparison Matrix - Test Scenarios

This document describes the comparison scenarios included in the fixtures for testing and demonstration.

## Laptop Comparison (Specification ID: 1)

Access via: `/quotes/compare/specification/1`

### Products Included

| Product | Brand | Compliance | Key Attributes | Price Range | Scenario |
|---------|-------|------------|----------------|-------------|----------|
| **MacBook Pro 15"** | Apple | ✓ 100% | 36GB RAM, 512GB SSD, 15.3", 12 cores | $2,399-2,499 | Premium option, complete specs, highest price |
| **Dell XPS 15** | Dell | ✓ 100% | 32GB RAM, 1024GB SSD, 15.6", 8 cores, Touchscreen | $1,649-1,699 | High-end option, complete specs, has touchscreen |
| **ThinkPad X1 Carbon** | Lenovo | ✓ 100% | 16GB RAM, 512GB SSD, 14", 10 cores | $1,299-1,399 | Business option, complete specs, smaller 14" display |
| **HP EliteBook 850** | HP | ⚠ 67% | 16GB RAM, **NO STORAGE SIZE**, 15", 8 cores | $1,099-1,149 | **INCOMPLETE** - Missing required Storage attribute |
| **Dell Latitude 5530** | Dell | ✓ 100% | 16GB RAM, 256GB SSD, 15", 6 cores | $899-979 | Budget option, complete specs, lowest price |

### Key Observations

1. **Cheapest Complete Option**: Dell Latitude 5530 at $899
   - 100% compliant, meets all required attributes
   - Lower specs (6 cores, 256GB) but best value

2. **Incomplete but Tempting**: HP EliteBook 850 at $1,099
   - Missing required "Storage" attribute (67% compliant)
   - Yellow background warning in comparison matrix
   - Cheaper than ThinkPad but incomplete data
   - **Decision Point**: Is it worth investigating vs. buying complete Latitude?

3. **Premium Options**: MacBook Pro and Dell XPS
   - Both 100% compliant
   - MacBook has more RAM (36GB vs 32GB)
   - Dell XPS has more storage (1TB vs 512GB) and touchscreen
   - Dell XPS is $750-850 cheaper than MacBook

4. **Multiple Vendors**: See price variations
   - Amazon often cheapest for single units
   - CDW offers bulk discounts (5+ units)
   - Best Buy retail prices highest

## Monitor Comparison (Specification ID: 2)

Access via: `/quotes/compare/specification/2`

### Products Included

| Product | Brand | Compliance | Key Attributes | Price Range | Scenario |
|---------|-------|------------|----------------|-------------|----------|
| **Dell UltraSharp U2720Q** | Dell | ✓ 100% | 27", 3840x2160, 60Hz, IPS | $649.99 | Professional option, ACCEPTED quote |
| **LG 27UK850-W** | LG | ✓ 100% | 27", 3840x2160, 60Hz, IPS | $599-629 | Nearly identical specs, cheaper |
| **Samsung M7** | Samsung | ✓ 100% | 32", 3840x2160, VA panel | $499-529 | Larger display, missing optional Refresh Rate |

### Key Observations

1. **Identical Specs, Different Prices**: Dell vs LG
   - Both 27", 4K, 60Hz, IPS panels
   - LG is $20-50 cheaper
   - Dell quote already ACCEPTED (status indicator)

2. **Samsung M7**: Different value proposition
   - Larger 32" display
   - Missing optional "Refresh Rate" attribute (still 100% compliant)
   - VA panel instead of IPS
   - Cheapest option at $499

## Comparison Features Demonstrated

### 1. Attribute Compliance Tracking
- **Green ✓ 100%**: All required attributes present
- **Orange ⚠ XX%**: Missing required attributes with score
- **Hover tooltip**: Shows which attributes are missing

### 2. Sorting
- Quotes sorted by price (lowest first)
- Makes cost comparison immediate

### 3. Visual Indicators
- **Yellow background**: Products with missing required attributes
- **Color-coded compliance**: Green (complete) vs Orange (incomplete)
- **Bold for required**: Required attributes marked with red asterisk (*)

### 4. Attribute Display
- **Specification attributes**: Core attributes defined in specification
- **Extra attributes toggle**: Show/hide product-specific extra features
- **Missing values**: Grayed out with dash (-)
- **Different data types**: Numbers, text, booleans all displayed appropriately

### 5. Multiple Quotes Per Product
- Dell Latitude: 3 different vendors ($899, $949, $979)
- Dell XPS 15: 3 active quotes ($1,649, $1,699, superseded)
- Shows vendor pricing competition

## Testing Recommendations

### Test Case 1: Complete vs Incomplete Decision
Navigate to `/quotes/compare/specification/1` and observe:
- HP EliteBook at $1,099 (incomplete, yellow warning)
- Dell Latitude at $899 (complete, lower price anyway)
- ThinkPad at $1,299 (complete, but more expensive than HP)

**Question**: Is the HP EliteBook worth investigating despite incomplete data?

### Test Case 2: Identical Specs Comparison
Navigate to `/quotes/compare/specification/2`:
- Dell and LG monitors have identical specs
- $50 price difference
- One quote already accepted

**Question**: Should the accepted Dell quote be reconsidered given LG's lower price?

### Test Case 3: Vendor Competition
Look at Dell Latitude 5530 quotes:
- CDW: $899 (5-unit minimum)
- Amazon: $949 (single unit)
- Best Buy: $979 (single unit)

**Question**: Is CDW's bulk discount worth committing to 5 units?

### Test Case 4: Extra Attributes
Toggle "Show extra attributes" on laptop comparison:
- See CPU Cores (optional attribute)
- MacBook: 12 cores
- Dell XPS: 8 cores
- ThinkPad: 10 cores

**Question**: Does core count matter for your use case?

## Data Integrity Examples

### Complete Products (100% Compliance)
- All required attributes have values
- Green checkmark indicator
- Safe to purchase

### Incomplete Products (<100% Compliance)
- HP EliteBook 850: Missing "Storage" (required)
  - Has RAM (16GB) ✓
  - Has Screen Size (15") ✓
  - **Missing Storage** ✗
  - Has Storage Type (SSD) - but no capacity!
  - Compliance: 67% (2 of 3 required attributes)

This demonstrates real-world scenario where vendors provide partial information, and the system helps identify data gaps before purchasing.

## API Endpoints

### Specification-Level Comparison
```
GET /quotes/compare/specification/:specId?show_extra=on
```
Example: `/quotes/compare/specification/1?show_extra=on`

### Product-Level Comparison
```
GET /quotes/compare/product/:productId?show_extra=on
```
Example: `/quotes/compare/product/2` (Dell XPS quotes from different vendors)

## Command-Line Access

List all laptop quotes:
```bash
./bin/buyer list quotes | grep -E "(MacBook|Dell|ThinkPad|HP|Latitude)"
```

View specific product:
```bash
./bin/buyer list products --limit 1 --offset 9  # HP EliteBook
```

## Summary

The fixtures provide realistic scenarios for:
- **Price comparison**: $899 to $2,499 range for laptops
- **Attribute compliance**: 67% to 100% scores
- **Vendor competition**: Multiple quotes per product
- **Data quality issues**: Incomplete specifications
- **Purchase decisions**: Cost vs completeness trade-offs

Use these scenarios to demonstrate the quote comparison matrix's value in procurement decision-making.
