# Pattern: Context Management & Progressive Disclosure

This document describes how we manage the agent's context window through progressive disclosure to maximize accuracy and token efficiency.

## Context
As codebases scale, feeding the entire repository into an AI context window results in context bloat, increased latency, token waste, and model hallucinations. We need a structured way to disclose information incrementally depending on the scope of the active task.

---

## Standard

### 1. Progressive Disclosure Flow
AI agents and developers should access project documentation in a structured pyramid:

1. **Orientation (The "What" and "Where")**: Look at `01_orientation/` first to establish a high-level conceptual mapping of the project structure and setup instructions.
2. **Patterns (The "How")**: Consult `02_patterns/` to understand the standard practices (coding conventions, mocking/testing standards, and alignment guidelines) before altering code.
3. **Deep Dives (The "Why" and "Detailed Mechanics")**: Read `03_deep_dives/` to understand specific, complex subsystems involved in the task. Do not load these unless your current task touches these components.
4. **Plans (The "State")**: Utilize `04_plans/` to track immediate steps, epic progress, and current/future goals.

### 2. File Size and Granularity Limits
- Document files should be focused on one conceptual unit.
- If a deep dive becomes larger than 300 lines, split it into smaller, targeted sub-documents.
- Keep standard code explanations inside documentation, keeping code comments sparse, focus comments on "why" rather than "what".

---

## Why
- **Saves Tokens**: Progressive disclosure avoids uploading unnecessary lines of code or tangential files.
- **Minimizes Regressions**: High conceptual isolation ensures that changing one component does not inadvertently corrupt assumptions about a neighboring component.
- **Speeds Up Onboarding**: New developers or agents can get to work within minutes by following the orientation guidelines without reading the entire codebase first.
