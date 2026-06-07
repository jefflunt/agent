# Task 2: Update Config and End-to-End Verification

## Objective
Update the global user configuration with the new `test_antigravity` adapter, and verify that the `agent` CLI behaves perfectly end-to-end.

## Steps
1. Append the following adapter to `~/.agent/config.yml`:
   ```yaml
   test_antigravity: "agy:google/gemini-3.5-flash"
   ```
2. Execute a real prompt through the `agent` CLI using the `test_antigravity` adapter:
   ```bash
   echo "What is 2+2?" | go run cmd/agent/main.go --verbose test_antigravity
   ```
3. Verify the output is clean and correct (e.g. prints `2 + 2 = 4`).
