# Project Architecture & Workflow

## Project Overview
Repo-lyzer follows a modular architecture designed for scalability and high-performance terminal rendering.

### Directory Structure
```text
repo-analyzer/
├── cmd/                # Cobra CLI commands
├── internal/
│   ├── github/         # API Client & Rate Limiting
│   ├── analyzer/       # Logic for Health/Maturity scores
│   ├── ui/             # Bubble Tea components & Lipgloss styles
│   └── config/         # Local settings persistence
└── docs/               # Technical documentation
```

### Core Workflow
**1. Trigger:** User executes a command via Cobra.
**2. Fetch:** The `github` package retrieves data with local caching.
**3. Analyze:** The `analyzer` package computes metrics (Bus Factor, Maturity).
**4. Render:** `Bubble Tea` manages the TUI state and `Lipgloss` handles the "Neon" styling.
