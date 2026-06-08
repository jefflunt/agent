package version

// Version holds the current compiled version of the agent CLI.
// It is injected at build time using -ldflags.
var Version = "dev"
