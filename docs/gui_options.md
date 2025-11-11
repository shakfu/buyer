# GUI Options for Go Applications

This document outlines the available GUI frameworks for embedding graphical interfaces in Go applications.

## 1. Fyne (Most Popular Cross-Platform)

**Description**: Pure Go GUI toolkit with Material Design aesthetics.

**Example**:
```go
import "fyne.io/fyne/v2/app"
import "fyne.io/fyne/v2/widget"

a := app.New()
w := a.NewWindow("Hello")
w.SetContent(widget.NewLabel("Hello Fyne!"))
w.ShowAndRun()
```

**Pros**:
- Pure Go implementation
- Cross-platform (Windows/Mac/Linux/Mobile)
- Material Design out of the box
- Active development and community
- Good documentation

**Cons**:
- Larger binaries (~15-30MB)
- Custom rendering (not native widgets)

**Best for**: Desktop applications, mobile apps, tools with traditional GUI needs

**Website**: https://fyne.io

---

## 2. Wails (Electron Alternative)

**Description**: Build desktop applications using Go backend with web frontend (HTML/CSS/JS).

**Example**:
```go
package main

import (
    "embed"
    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := &App{}

    err := wails.Run(&options.App{
        Title:  "Buyer",
        Width:  1024,
        Height: 768,
        Assets: assets,
        Bind: []interface{}{
            app,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

**Pros**:
- Use existing web technologies (HTML/CSS/JavaScript)
- Native feel (uses system webview)
- Smaller than Electron (~5-10MB vs 100MB+)
- Can reuse existing web UI code
- Live development mode
- TypeScript bindings auto-generated

**Cons**:
- Requires WebView2 on Windows (usually pre-installed on Windows 11)
- Web-based limitations apply

**Best for**: Converting web applications to desktop, apps with complex UIs, developers familiar with web tech

**Perfect for Buyer app**: Since we already have Pico CSS + HTMX web interface!

**Website**: https://wails.io

---

## 3. Gio (Immediate Mode GUI)

**Description**: Low-level immediate mode GUI framework, inspired by Dear ImGui.

**Example**:
```go
package main

import (
    "gioui.org/app"
    "gioui.org/io/system"
    "gioui.org/layout"
    "gioui.org/widget/material"
)

func main() {
    go func() {
        w := app.NewWindow()
        th := material.NewTheme()
        var ops op.Ops
        for e := range w.Events() {
            if e, ok := e.(system.FrameEvent); ok {
                gtx := layout.NewContext(&ops, e)
                // Draw UI here
                e.Frame(gtx.Ops)
            }
        }
    }()
    app.Main()
}
```

**Pros**:
- Very small binaries
- Immediate mode rendering
- Full control over rendering
- Good performance
- Pure Go

**Cons**:
- Lower-level API (more code needed)
- Steeper learning curve
- Less "out of the box" functionality

**Best for**: Games, real-time visualization tools, performance-critical apps

**Website**: https://gioui.org

---

## 4. go-app (Progressive Web Apps)

**Description**: Framework for building Progressive Web Apps using Go and WebAssembly.

**Example**:
```go
package main

import (
    "github.com/maxence-charriere/go-app/v9/pkg/app"
)

type hello struct {
    app.Compo
}

func (h *hello) Render() app.UI {
    return app.H1().Text("Hello World!")
}

func main() {
    app.Route("/", &hello{})
    app.RunWhenOnBrowser()

    http.Handle("/", &app.Handler{
        Name:        "Hello",
        Description: "An Hello World! example",
    })
    http.ListenAndServe(":8000", nil)
}
```

**Pros**:
- Write Go code that runs in browser (WebAssembly)
- Works offline (PWA features)
- Installable as "app" on desktop/mobile
- Familiar web paradigms

**Cons**:
- Still browser-based
- WebAssembly limitations
- Larger initial download

**Best for**: Web apps that need offline capability, apps targeting both web and "desktop"

**Website**: https://go-app.dev

---

## 5. Other Notable Options

### walk (Windows Native)
- Windows-only
- Native Windows controls
- Lightweight
- https://github.com/lxn/walk

### ui (libui bindings)
- Native widgets on each platform
- Minimal
- Less actively maintained
- https://github.com/andlabs/ui

### Muon (Electron-like)
- Uses Chrome Embedded Framework
- Similar to Electron
- Larger binaries
- https://github.com/ImVexed/muon

---

## Recommendation for Buyer App

**Best Choice: Wails v2**

### Reasoning:
1. **Reuse Existing UI**: Already have Pico CSS + HTMX web interface
2. **Native Feel**: Uses system webview (not Chromium bundled)
3. **Small Binary**: ~5-10MB final size
4. **Easy Migration**: Minimal changes to existing code
5. **Rich Features**: System dialogs, notifications, file system access
6. **Active Development**: Well-maintained, good docs

### Migration Path:
1. Keep existing Go backend (services, models)
2. Move web templates to Wails frontend directory
3. Add Wails runtime to expose Go functions to JavaScript
4. Build as native desktop app
5. Optional: Keep web server mode for remote access

### Benefits:
- Single binary distribution
- No browser required
- System tray integration possible
- Native file/folder dialogs
- Better offline experience
- Professional desktop feel

---

## Quick Comparison Table

| Framework | Binary Size | Learning Curve | Cross-Platform | Native Look | Web Tech |
|-----------|-------------|----------------|----------------|-------------|----------|
| Fyne      | 15-30MB     | Medium         | [x] Full        | Custom      | [X]       |
| Wails     | 5-10MB      | Low*           | [x] Full        | [x] Native   | [x]       |
| Gio       | <5MB        | High           | [x] Full        | Custom      | [X]       |
| go-app    | Variable    | Medium         | [x] Web         | Web         | [x]       |

*Low if already familiar with web development

---

## Getting Started with Wails (Quick Start)

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Create new project
wails init -n buyer-desktop -t vanilla

# Or convert existing project
wails init -n buyer-desktop -t vanilla-lite

# Develop with live reload
wails dev

# Build for production
wails build
```

---

## Resources

- **Fyne**: https://developer.fyne.io/
- **Wails**: https://wails.io/docs/introduction
- **Gio**: https://gioui.org/doc
- **go-app**: https://go-app.dev/start
- **Awesome Go GUIs**: https://github.com/avelino/awesome-go#gui
