# Task 2: End-to-End Verification with Gemini CLI

## Objective
Verify that the `agent` CLI behaves perfectly end-to-end when using the `test_gemini` adapter.

## Steps
1. Execute a real prompt through the `agent` CLI using the `test_gemini` adapter:
   ```bash
   echo "What is 2+2?" | go run cmd/agent/main.go --verbose test_gemini
   ```
2. Verify the output is clean and correct (e.g. prints `4`).
