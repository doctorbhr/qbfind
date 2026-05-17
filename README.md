# QBFind: High-Performance Windows Desktop File Finder

[![Go Version](https://img.shields.io/github/go-mod/go-version/tommyvercetti89/qbfind?logo=go)](https://golang.org)
[![License](https://img.shields.io/github/license/tommyvercetti89/qbfind)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows-blue?logo=windows)](https://microsoft.com)

**QBFind** is a professional-grade, high-performance desktop file finder built specifically for Microsoft Windows. To cater to both retro-performance enthusiasts and lovers of modern, sleek UI aesthetics, QBFind is distributed in **two distinct, specialized editions**:

---

## 💎 Dual-Edition Architecture

| Feature | 🚀 QBFind Classic (Win32 Edition) | ✨ QBFind Premium (Wails Edition) |
| :--- | :--- | :--- |
| **User Interface** | Pure Native Win32 Controls (SysListView32, Edit, Static) | Modern Pitch-Black (#000000) & Apple-style Light Theme |
| **Tech Stack** | Pure Go, Raw Windows DLL APIs, Win32 Message Loop | Go Backend, Wails Runtime, Vite, HTML5, Vanilla CSS3, JS |
| **Executable Size**| Ultra-lightweight (~3.4 MB) | Compact Premium Desktop App (~12 MB) |
| **Memory / CPU** | Near-zero footprint, instant cold-starts | Low-footprint GPU-accelerated WebView2 rendering |
| **Key Highlights** | Manual COM/OLE drag-and-drop, raw Win32 windows | Cam-efektli (glassmorphic) sağ tık menüleri, akışkan mikro-animasyonlar, yerelleştirilmiş araç ipuçları (tooltips) |

---

## ⚡ Core Shared Features

*   **⚡ Parallel Multi-Threaded Scanning**: Implements a high-performance concurrent worker pool dynamically sized based on CPU cores:
    $$\text{Workers} = \min(\max(\text{NumCPU} \times 2, 2), 16)$$
    Crawls userspace paths (`Desktop`, `Downloads`, `Documents`, etc.) first to ensure instantaneous query availability upon startup.
*   **🎯 Intelligent Scoring & Ranking Engine**: Sorts search results using a comprehensive ceza/ödül (penalty/reward) scoring algorithm based on path depth, folder priority, executable extensions, and Turkish character folding (`ı, İ, ç, ğ, ö, ş, ü` mappings).
*   **🔍 In-App Text Preview**: Decodes UTF-8, UTF-16LE, and UTF-16BE files dynamically with Byte Order Mark (BOM) validation, displaying raw text previews up to 4,000 characters.
*   **🌍 Multilingual Support**: Real-time language switching between English and Turkish, with persistent configurations saved under `%APPDATA%/QBFind/settings.txt`.

---

## 🏗️ Architecture & Workflows

### 1. Win32 Classic Message Loop
```mermaid
graph TD
    A[main.go - LockOSThread] --> B[DLL Procedures & COM Initializers]
    B --> C[RegisterClassExW & CreateWindowExW]
    C --> D[WndProc Event Router - GetMessageW]
    D -->|WM_CREATE| E[Initialize native ListView & Controls]
    D -->|WM_CREATE| F[Background Indexer Goroutine]
    D -->|WM_COMMAND| G[Actions Manager: Open, Preview, Copy Path, Info]
    D -->|WM_TIMER| H[Debounced Search Engine]
    D -->|WM_NOTIFY| I[ListView Context Menu / Drag & Drop]
    F -->|Worker Pool| J[Drive Crawler -> In-Memory Cache]
```

### 2. Wails Modern IPC Architecture
```mermaid
graph LR
    subgraph Frontend [Vite Webview2 Layer]
        A[index.html / CSS Variables] <--> B[main.js Event Listeners]
        B <--> C[Right-Click Context Menu & Tooltips]
    end
    subgraph Backend [Go Native Layer]
        D[app.go - IPC Bindings] <--> E[search.go - Normalizer & Scorer]
        E <--> F[indexer.go - Multi-Threaded Scanner]
    end
    B <-->|Wails IPC Bindings / Events| D
```

---

## 📂 Modular File Walkthrough

### 🚀 Classic Win32 Codebase (Root Directory)
*   **`quickfind.go`**: Entrypoint (`main`), DLL procedure bindings, Win32 structs, and global variables.
*   **`ui.go`**: Native `wndProc` routing, child controls creation, layout constraints (`layoutControls`), and dialog popups.
*   **`listview.go`**: Column definitions, double-click/right-click handlers, and native row insertion.
*   **`search.go`**: Debounced search trigger, Turkish text normalization (`searchFold`), and scoring/penalty algorithms (`scoreEntry`).
*   **`indexer.go`**: Concurrent logical drive scanning worker pool, junction/reparse point validation, and folder prioritizer.
*   **`ole.go`**: Low-level COM interface simulation (VTables) for native Windows Drag-and-Drop (`DoDragDrop`).
*   **`settings.go`**: Persistent configurations cache and English/Turkish translation registers.
*   **`actions.go`**: Native execution triggers (`ShellExecuteW`), desktop copy operations (`SHFileOperationW`), and metadata previews.
*   **`utils.go`**: String transformations, numeric formatting, clipboard writers, and bitwise macros.

### ✨ Premium Wails Codebase (`/wailsapp`)
*   **`wailsapp/main.go`**: Go entrypoint setting up the Wails application instance, options, and windows size constraints.
*   **`wailsapp/app.go`**: Cross-compiled Go handlers bound to the frontend (search pipelines, clipboard, open file wrappers, scanning).
*   **`wailsapp/indexer.go`**: Thread-safe parallel index crawler emitting event payloads (`status_update`) back to JS.
*   **`wailsapp/search.go`**: Token-matching and scoring engines optimized for Wails' memory data-bindings.
*   **`wailsapp/frontend/`**: The modern Webview2 UI stack powered by Vite, HTML5, Vanilla CSS3 (variables custom theme support), and Vanilla JavaScript modules.

---

## 🛠️ Compilation & Development Guides

Ensure you have **Go 1.18 or higher** installed. To compile the Premium Edition, you will also need the **Wails CLI** installed (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).

### 1. Developing & Building Classic Win32 Edition
```powershell
# Run in development mode
go run .

# Compile optimized production standalone .exe (without console shell window)
go build -ldflags "-s -w -H windowsgui" -o QBFind_Classic.exe
```

### 2. Developing & Building Premium Wails Edition
```powershell
# Navigate to Wails app
cd wailsapp

# Run in live-development mode (with hot-reloads)
wails dev

# Compile optimized production standalone .exe
wails build -o QBFind_Premium.exe -clean -trimpath -ldflags "-s -w"
```

---

## 🤝 Contribution Guidelines

We welcome contributions to both the Win32 Classic and Wails Premium editions! To maintain consistency:

1.  **Strict Modularization**: Keep Win32 API logic in the root directory and Wails-specific wrappers under the `/wailsapp` directory. Do not mix dependencies.
2.  **No Unnecessary Dependencies**: Classic edition must remain 100% Cgo-free and rely solely on raw system DLLs. Premium edition must use standard Vanilla CSS variables to ensure zero tailwind/unnecessary node-module overhead.
3.  **Bilingual Support**: Always implement localization for both English (`en`) and Turkish (`tr`) for any new strings, context menus, tooltips, or alerts.

---

## 🤝 Collaboration Acknowledgement

Both editions of this high-performance suite were designed, modularized, and refined through a pair-programming partnership with **Antigravity**, an advanced agentic AI coding assistant designed by the **Google DeepMind** team.

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
