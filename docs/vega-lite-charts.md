# Using Vega-Lite for Visualizations

The Buyer application includes Vega, Vega-Lite, and Vega-Embed libraries for creating interactive data visualizations.

## Libraries Included

Located in `/cmd/buyer/web/static/js/`:
- `vega.min.js` - The core Vega visualization grammar
- `vega-lite.min.js` - High-level grammar for statistical graphics
- `vega-embed.min.js` - Embed Vega visualizations in web pages

These libraries are automatically loaded in the base template.

## Basic Usage

### 1. Simple Bar Chart

```html
<div id="my-chart"></div>

<script>
const data = [
    {category: "A", value: 28},
    {category: "B", value: 55},
    {category: "C", value: 43}
];

const spec = {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "width": "container",
    "height": 300,
    "data": {"values": data},
    "mark": "bar",
    "encoding": {
        "x": {"field": "category", "type": "nominal", "title": "Category"},
        "y": {"field": "value", "type": "quantitative", "title": "Value"},
        "tooltip": [
            {"field": "category", "type": "nominal"},
            {"field": "value", "type": "quantitative"}
        ]
    }
};

vegaEmbed('#my-chart', spec, {actions: false});
</script>
```

### 2. Line Chart

```javascript
const timeSeriesSpec = {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "width": "container",
    "height": 300,
    "data": {"values": data},
    "mark": "line",
    "encoding": {
        "x": {"field": "date", "type": "temporal", "title": "Date"},
        "y": {"field": "price", "type": "quantitative", "title": "Price (USD)"},
        "tooltip": [
            {"field": "date", "type": "temporal", "format": "%Y-%m-%d"},
            {"field": "price", "type": "quantitative", "format": "$,.2f"}
        ]
    }
};
```

### 3. Scatter Plot

```javascript
const scatterSpec = {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "width": "container",
    "height": 400,
    "data": {"values": data},
    "mark": "point",
    "encoding": {
        "x": {"field": "quantity", "type": "quantitative"},
        "y": {"field": "price", "type": "quantitative"},
        "color": {"field": "vendor", "type": "nominal"},
        "size": {"field": "total", "type": "quantitative"},
        "tooltip": [
            {"field": "vendor", "type": "nominal"},
            {"field": "quantity", "type": "quantitative"},
            {"field": "price", "type": "quantitative", "format": "$,.2f"}
        ]
    }
};
```

### 4. Grouped Bar Chart

```javascript
const groupedBarSpec = {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "width": "container",
    "height": 300,
    "data": {"values": data},
    "mark": "bar",
    "encoding": {
        "x": {"field": "product", "type": "nominal"},
        "y": {"field": "price", "type": "quantitative"},
        "color": {"field": "vendor", "type": "nominal"},
        "xOffset": {"field": "vendor"}
    }
};
```

## Using with Go Templates

When embedding data from Go templates:

```html
<script>
const chartData = [
    {{range .DataPoints}}
    {
        name: "{{.Name}}",
        value: {{.Value}},
        category: "{{.Category}}"
    },
    {{end}}
];

const spec = {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "data": {"values": chartData},
    "mark": "bar",
    "encoding": {
        "x": {"field": "name", "type": "nominal"},
        "y": {"field": "value", "type": "quantitative"}
    }
};

vegaEmbed('#chart', spec);
</script>
```

## Common Encodings

### Temporal (Time)
- Use `"type": "temporal"` for dates
- Format with `"format": "%Y-%m-%d"`

### Quantitative (Numbers)
- Use `"type": "quantitative"` for continuous numbers
- Format currency: `"format": "$,.2f"`
- Format percentages: `"format": ".1%"`

### Nominal (Categories)
- Use `"type": "nominal"` for discrete categories
- Good for vendor names, product types, etc.

### Ordinal (Ordered Categories)
- Use `"type": "ordinal"` for ordered discrete data
- Good for priority levels, size categories

## Color Schemes

Popular color schemes:
- `"blues"` - Sequential blue scale
- `"viridis"` - Perceptually uniform
- `"category10"` - 10 categorical colors
- `"redyellowblue"` - Diverging scale

```javascript
"color": {
    "field": "value",
    "type": "quantitative",
    "scale": {"scheme": "blues"}
}
```

## Interactive Features

### Tooltips
Always include tooltips for better UX:

```javascript
"tooltip": [
    {"field": "vendor", "type": "nominal", "title": "Vendor"},
    {"field": "price", "type": "quantitative", "title": "Price", "format": "$,.2f"}
]
```

### Selection (Click/Hover)
```javascript
"selection": {
    "highlight": {"type": "single", "on": "mouseover"}
},
"encoding": {
    "opacity": {
        "condition": {"selection": "highlight", "value": 1},
        "value": 0.5
    }
}
```

## vegaEmbed Options

```javascript
vegaEmbed('#chart', spec, {
    actions: false,        // Hide action menu
    theme: 'latimes',      // Use LA Times theme
    renderer: 'canvas',    // Use canvas (default is svg)
    tooltip: true          // Enable tooltips
});
```

## Real-World Examples in Buyer

### Dashboard: Vendor Spending Chart
See `/cmd/buyer/web/templates/dashboard.html` for a complete example of:
- Loading data from Go templates
- Creating a bar chart
- Adding tooltips
- Color encoding by value

## Resources

- [Vega-Lite Documentation](https://vega.github.io/vega-lite/docs/)
- [Example Gallery](https://vega.github.io/vega-lite/examples/)
- [Vega-Embed Options](https://github.com/vega/vega-embed#options)
- [Online Editor](https://vega.github.io/editor/)

## Tips

1. **Always set width to "container"** for responsive charts
2. **Use tooltips** - they greatly improve usability
3. **Test with real data** - edge cases (empty data, large datasets) matter
4. **Keep it simple** - Vega-Lite is designed for clarity
5. **Format numbers** - Use format strings for currency, percentages, etc.
