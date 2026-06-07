# 04 Plans

This directory contains actionable work items, epics, implementation approaches, and active feature design plans.

## Active & Historical Plans

- [x] **`bootstrap-agent-cli`**: Establishing the initial centralized Go-based `agent` CLI tool.
- [x] **`implement-claude-driver`**: Designing and integrating the Claude Code (`claude`) driver.
- [x] **`implement-gemini-driver`**: Designing and integrating the Gemini CLI (`gemini`) driver.
- [x] **`implement-agy-driver`**: Designing and integrating the Google Antigravity (`agy`) driver.

---

## Guidelines for Proposing Plans

1. **Create an Epic Folder**: Group designs and state files inside a dedicated folder under `agent_docs/04_plans/<epic-name>/`.
2. **Draft a Design Specification**: Create a `design.md` based on `agent_docs/templates/plan_template.md`.
3. **Review and Align**: Align on the approach (Continuous Alignment) before writing code.
4. **Track State**: Update the plan file checklist steps from `Pending` (`- [ ]`) to `Done` (`- [x]`) as tasks are implemented.
