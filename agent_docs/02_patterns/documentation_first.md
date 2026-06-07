# Pattern: Documentation-First

This document outlines the standard for treating documentation as the project's technical contract.

## Context
When collaborating across distributed teams or working with AI agents, misalignments often arise due to diverging assumptions about system requirements, API boundaries, or architectural decisions. Documentation-First practices eliminate this friction by creating a verifiable blueprint before any line of code is modified.

## Standard

### 1. Plan Before Implementing
- Prior to launching any new feature, refactoring, or driver addition, a plan file must be drafted in `agent_docs/04_plans/<epic_name>/design.md` or a similar path.
- Outline clear objectives, scope limitations, detailed breakdown steps, verification procedures, and risk mitigations.

### 2. State-Tracking Mechanics
- Each step of the plan is tracked sequentially.
- When an agent/developer works on a step, they update its status.
- Once completed, the step is marked as done in the plan file.

### 3. Update Documentation Concurrently
- Do not let documentation go stale. If an implementation details deviates from the original design during development, update the architectural context (`01_orientation`) or patterns (`02_patterns`) *before* committing the code.

---

## Why
- **Verifiability**: Clear goals and step-by-step verification methods allow agents to self-correct during the development loop.
- **Context Management**: It keeps instructions bounded, minimizing context drift and avoiding code regressions.
- **Contractual Alignment**: It establishes a clear contract of what is being built, making code reviews and integration testing straightforward.
