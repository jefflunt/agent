# Implement a Go package 'pkg/adapter' that parses a configuration target string formatted as 'cliName:provider/model' into distinct, accessible components (CLIName, Provider, and Model), returning formatted, standard error values for any malformed input formats.

The objective of this task is to implement the adapter specification parser in the 'pkg/adapter' package of the Go-based 'agent' CLI tool.

This parser is responsible for taking a target string like 'opencode:google/gemini-3.5-flash' from the configurations loaded by 'pkg/config' and breaking it down into its constituent parts: the runner command/CLI name (e.g., 'opencode'), the model provider (e.g., 'google'), and the model name itself (e.g., 'gemini-3.5-flash').

It must validate that the incoming target format matches the expected 'cliName:provider/model' structure. If the target string is malformed or missing any of these parts, the parser should return standard, well-typed errors. The implementation must be designed defensively to handle edge cases, such as trailing or leading spaces, empty components, and missing or extra separators.
