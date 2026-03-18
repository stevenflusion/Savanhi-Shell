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
| When structuring Java apps by Domain/Application/Infrastructure, or refactoring toward clean architecture | hexagonal-architecture-layers-java | /home/steven/.config/opencode/skills/hexagonal-architecture-layers-java/SKILL.md |
| During Elixir code review, refactoring sessions, or when writing Phoenix/Ecto code | elixir-antipatterns | /home/steven/.config/opencode/skills/elixir-antipatterns/SKILL.md |
| When building desktop apps, working with Electron main/renderer processes, IPC communication, or native integrations | electron | /home/steven/.config/opencode/skills/electron/SKILL.md |
| When building mobile apps, working with React Native components, using Expo, React Navigation, or NativeWind | react-native | /home/steven/.config/opencode/skills/react-native/SKILL.md |
| When user asks to create a Jira task, ticket, or issue | jira-task | /home/steven/.config/opencode/skills/jira-task/SKILL.md |
| When user asks to create an epic, large feature, or multi-task initiative | jira-epic | /home/steven/.config/opencode/skills/jira-epic/SKILL.md |

## Project Conventions

No project convention files found. This is a new project with only PRD.md present.

## Stack Detection

Based on PRD.md:
- **Primary Language**: Go 1.21+
- **TUI Framework**: Bubble Tea (Charm)
- **Styling**: Lipgloss
- **Configuration**: JSON
- **Target**: Terminal ecosystem configurator with real-time preview

## Relevant Skills for This Project

| Skill | Relevance |
|-------|-----------|
| **go-testing** | HIGH - Go testing patterns including Bubbletea TUI testing with teatest |
| skill-creator | MEDIUM - May need to create project-specific skills |
| github-pr | LOW - For PR workflow when contributing |

## Notes

- Project is in early stage (PRD only)
- Primary skill: `go-testing` for TUI testing with Bubbletea/teatest
- No linters, test frameworks, or CI configuration detected yet