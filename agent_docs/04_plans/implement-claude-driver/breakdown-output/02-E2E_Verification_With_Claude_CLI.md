# Task 2: End-to-End Verification with Claude CLI

## Objective
Verify that the `agent` CLI behaves perfectly end-to-end when using the `test_claude` adapter.

## Steps
1. Execute a real prompt through the `agent` CLI using the `test_claude` adapter:
   ```bash
   echo "What is 2+2?" | go run cmd/agent/main.go --verbose test_claude
   ```
2. Verify the output is clean and correct.
3. If any model availability issues arise (such as `claude-sonnet-4-6` not being supported by the user's Claude Code installation), we will note the behavior, coordinate model adjustments (e.g., using `claude-sonnet-4-5` or `haiku`), and ensure it succeeds.
