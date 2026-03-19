# Skill Registry

**Orchestrator use only.** Read this registry once per session to resolve skill paths, then pass pre-resolved paths directly to each sub-agent's launch prompt. Sub-agents receive the path and load the skill directly — they do NOT read this registry.

## User Skills

| Trigger | Skill | Path |
|---------|-------|------|
| When writing Go tests, using teatest, or adding test coverage | go-testing | /home/steven/.config/opencode/skills/go-testing/SKILL.md |
| When user asks to create a new skill, add agent instructions, or document patterns for AI | skill-creator | /home/steven/.config/opencode/skills/skill-creator/SKILL.md |
| When writing TypeScript code - types, interfaces, generics | typescript | /home/steven/.config/opencode/skills/typescript/SKILL.md |
| When writing Python tests - fixtures, mocking, markers | pytest | /home/steven/.config/opencode/skills/pytest/SKILL.md |
| When writing React components - no useMemo/useCallback needed | react-19 | /home/steven/.config/opencode/skills/react-19/SKILL.md |
| When creating PRs, writing PR descriptions, or using gh CLI for pull requests | github-pr | /home/steven/.config/opencode/skills/github-pr/SKILL.md |
| When working with Next.js - routing, Server Actions, data fetching | nextjs-15 | /home/steven/.config/opencode/skills/nextjs-15/SKILL.md |
| When managing React state with Zustand | zustand-5 | /home/steven/.config/opencode/skills/zustand-5/SKILL.md |
| When styling with Tailwind - cn(), theme variables, no var() in className | tailwind-4 | /home/steven/.config/opencode/skills/tailwind-4/SKILL.md |
| When using Zod for validation - breaking changes from v3 | zod-4 | /home/steven/.config/opencode/skills/zod-4/SKILL.md |
| When building AI chat features - breaking changes from v4 | ai-sdk-5 | /home/steven/.config/opencode/skills/ai-sdk-5/SKILL.md |
| When writing E2E tests - Page Objects, selectors, MCP workflow | playwright | /home/steven/.config/opencode/skills/playwright/SKILL.md |
| When building REST APIs with Django - ViewSets, Serializers, Filters | django-drf | /home/steven/.config/opencode/skills/django-drf/SKILL.md |
| When writing Java 21 code using records, sealed types, or virtual threads | java-21 | /home/steven/.config/opencode/skills/java-21/SKILL.md |
| When building or refactoring Spring Boot 3 applications | spring-boot-3 | /home/steven/.config/opencode/skills/spring-boot-3/SKILL.md |
| When structuring Java apps by Domain/Application/Infrastructure, or refactoring toward clean architecture | hexagonal-architecture-layers-java | /home/steven/.config/openerate/skills/hexagonal-architecture-layers-java/SKILL.md |
| During Elixir code review, refactoring sessions, or when writing Phoenix/Ecto code | elixir-antipatterns | /home/steven/.config/opencode/skills/elixir-antipatterns/SKILL.md |
| When building desktop apps, working with Electron main/renderer processes, IPC communication, or native integrations | electron | /home/steven/.config/opencode/skills/electron/SKILL.md |
| When building mobile apps, working with React Native components, using Expo, React Navigation, or NativeWind | react-native | /home/steven/.config/opencode/skills/react-native/SKILL.md |
| When user asks to create a Jira task, ticket, or issue | jira-task | /home/steven/.config/opencode/skills/jira-task/SKILL.md |
| When user asks to create an epic, large feature, or multi-task initiative | jira-epic | /home/steven/.config/opencode/skills/jira-epic/SKILL.md |

## Project Conventions

### Tech Stack
- **Language**: Go 1.24.2 (requires 1.21+)
- **TUI Framework**: Bubble Tea (charmbracelet/bubbletea v1.3.10)
- **Styling**: Lipgloss (charmbracelet/lipgloss v1.1.0)
- **UI Components**: Bubbles (charmbracelet/bubbles v1.0.0)
- **Testing**: teatest (charmbracelet/x/exp/teatest), standard go test with race detection
- **Build**: Make + goreleaser for cross-platform releases

### Architecture Patterns
- **TUI Pattern**: Bubble Tea with Screen enum for navigation (`internal/tui/model.go`)
- **Detection Pattern**: Detector interface with DetectAll() (`internal/detector/`)
- **Verification Pattern**: Verifier struct with VerifyComplete() (`internal/installer/verify.go`)
- **Persistence Pattern**: JSON files in ~/.config/savanhi/ (`internal/persistence/`)

### Code Style
- Standard Go formatting (gofmt)
- golangci-lint for linting
- Race detection in tests (`go test -race`)
- Coverage tracking with Codecov

### Key Directories
```
cmd/savanhi-shell/     # Entry point + CLI flags
internal/
├── cli/               # Non-interactive CLI
├── detector/          # OS/Shell/Terminal detection
├── errors/            # Error handling with exit codes
├── installer/         # Installation engine + Verifier
├── persistence/       # JSON preferences & history
├── preview/           # Live subshell preview
├── staging/           # Change staging system
└── tui/               # Bubble Tea interface
    ├── model.go       # Screen enum, Model struct
    ├── view.go        # Render methods per screen
    ├── update.go      # Key handlers per screen
    └── styles/        # Lipgloss ColorPalette & styles
pkg/shell/             # Public shell interface
configs/bundled/       # Bundled oh-my-posh themes
tests/e2e/             # End-to-end tests
```

### Adding New TUI Screen (Pattern)
1. Add constant to `Screen` enum in `internal/tui/model.go`
2. Add render method in `internal/tui/views/`: `renderNewScreen()`
3. Add key handler in `internal/tui/update.go`: `handleNewScreenKeys()`
4. Add cases in `View()` and `handleKeyPress()` switches
5. Add navigation from/to other screens as needed

### Testing TUI (Pattern)
- Use `teatest` for testing Bubble Tea models
- Test model updates: `model, cmd := Update(msg)`
- Test view output: `view := View()` string comparison

## Relevant Skills for This Project

| Skill | Relevance | When to Use |
|-------|-----------|-------------|
| **go-testing** | HIGH | All Go test files, TUI testing with teatest |
| skill-creator | MEDIUM | Creating project-specific skills for patterns |
| github-pr | LOW | PR workflow for contributions |

## SDD Context

SDD initialized with `engram` mode for artifact persistence.

**Engram Topic Key**: `sdd-init/savanhi-shell`

## Notes

- Project follows standard Go project layout
- TUI uses Bubble Tea's Elm Architecture (Model-Update-View)
- All existing screens follow the same pattern (model.go, view.go, update.go)
- Verifier in `internal/installer/verify.go` already has component verification logic to reuse for Health Dashboard
- CI: GitHub Actions with test matrix (Go 1.21/1.22, ubuntu/macos)