package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

// Config holds the parsed CLI flags.
type Config struct {
	ComposeFiles []string
	User         string
	Host         string
	KeyPath      string
	Command      string
}

// HelpText is the complete help output printed when --help is passed.
const HelpText = `ship — build, transfer, and deploy Docker Compose images to a remote host

Usage:
  ship [flags]

Required Flags:
  --docker-compose <path>   Path to compose file(s), comma-separated for multiple
  --user <user>             SSH user on the remote host
  --host <host>             Remote host address
  --key <path>              Path to SSH private key file
  --command <cmd>           Command to execute on the remote host after transfer

Examples:
  ship --docker-compose docker-compose.yml --user deploy --host 10.0.0.5 --key ~/.ssh/id_ed25519 --command "docker compose up -d"
  ship --docker-compose compose.yml,compose.prod.yml --user root --host staging.example.com --key ./key.pem --command "docker compose pull && docker compose up -d"
`

const usageLine = "Usage: ship --docker-compose <path> --user <user> --host <host> --key <path> --command <cmd>"

// splitComposeFiles splits a comma-separated string into a slice of trimmed paths.
// Returns nil if input is empty or contains only whitespace/commas.
func splitComposeFiles(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Parse parses CLI flags from args and returns a Config.
// Returns flag.ErrHelp when --help is requested.
// Returns an error listing all missing flags if any required flag is absent or empty.
func Parse(args []string) (Config, error) {
	fs := flag.NewFlagSet("ship", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var cfg Config
	var rawCompose string
	fs.StringVar(&rawCompose, "docker-compose", "", "")
	fs.StringVar(&cfg.User, "user", "", "")
	fs.StringVar(&cfg.Host, "host", "", "")
	fs.StringVar(&cfg.KeyPath, "key", "", "")
	fs.StringVar(&cfg.Command, "command", "", "")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	cfg.ComposeFiles = splitComposeFiles(rawCompose)

	// Track which flags were explicitly provided.
	explicitlySet := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		explicitlySet[f.Name] = true
	})

	var missing []string
	if len(cfg.ComposeFiles) == 0 {
		missing = append(missing, "--docker-compose")
	}
	if cfg.User == "" {
		missing = append(missing, "--user")
	}
	if cfg.Host == "" {
		missing = append(missing, "--host")
	}
	if cfg.KeyPath == "" {
		missing = append(missing, "--key")
	}

	commandEmpty := strings.TrimSpace(cfg.Command) == ""
	switch {
	case !explicitlySet["command"]:
		missing = append(missing, "--command")
	case commandEmpty && len(missing) == 0:
		return Config{}, fmt.Errorf("Empty --command flag — provide the command to run on the remote host") //nolint:staticcheck // user-facing message per DESIGN.md spec
	case commandEmpty:
		missing = append(missing, "--command")
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("Missing required flags: %s\n%s", strings.Join(missing, ", "), usageLine) //nolint:staticcheck // user-facing message per DESIGN.md spec
	}

	return cfg, nil
}
