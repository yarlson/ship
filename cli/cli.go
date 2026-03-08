package cli

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Config holds the parsed CLI arguments.
type Config struct {
	Image   string
	User    string
	Host    string
	KeyPath string
	Port    int
}

// HelpText is the complete help output printed when --help is passed.
const HelpText = `ship — transfer a local Docker image to a remote host over SSH

Usage:
  ship [flags] <user@host> <image[:tag]>

Flags:
  -i, --identity-file <path>  Path to SSH private key file
  -p, --port <port>           SSH port (default: 22)

Examples:
  ship root@10.0.0.1 app:latest
  ship -i ~/.ssh/id_ed25519 deploy@staging.example.com app:latest
  ship -i ~/.ssh/id_ed25519 -p 2222 deploy@staging.example.com ghcr.io/acme/app:dev
`

const usageLine = "Usage: ship [-i <path>] [-p <port>] <user@host> <image[:tag]>"

// Parse parses CLI args and returns a Config.
// Returns flag.ErrHelp when --help is requested.
func Parse(args []string) (Config, error) {
	fs := flag.NewFlagSet("ship", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var cfg Config
	fs.StringVar(&cfg.KeyPath, "i", "", "")
	fs.StringVar(&cfg.KeyPath, "identity-file", "", "")
	fs.IntVar(&cfg.Port, "p", 22, "")
	fs.IntVar(&cfg.Port, "port", 22, "")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	explicitlySet := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		explicitlySet[f.Name] = true
	})

	if (explicitlySet["i"] || explicitlySet["identity-file"]) && strings.TrimSpace(cfg.KeyPath) == "" {
		return Config{}, fmt.Errorf("empty -i flag — provide the path to an SSH private key")
	}

	if cfg.Port <= 0 {
		return Config{}, fmt.Errorf("invalid -p value: %s — port must be greater than 0", strconv.Itoa(cfg.Port))
	}

	positional := fs.Args()
	missing := missingPositionals(positional)
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required arguments: %s\n%s", strings.Join(missing, ", "), usageLine)
	}

	if len(positional) > 2 {
		return Config{}, fmt.Errorf("unexpected arguments: %s\n%s", strings.Join(positional[2:], ", "), usageLine)
	}

	user, host, err := parseTarget(positional[0])
	if err != nil {
		return Config{}, err
	}

	cfg.User = user
	cfg.Host = host
	cfg.Image = positional[1]

	return cfg, nil
}

func missingPositionals(positional []string) []string {
	switch len(positional) {
	case 0:
		return []string{"<user@host>", "<image[:tag]>"}
	case 1:
		return []string{"<image[:tag]>"}
	default:
		return nil
	}
}

func parseTarget(target string) (user, host string, err error) {
	user, host, ok := strings.Cut(target, "@")
	if !ok || strings.TrimSpace(user) == "" || strings.TrimSpace(host) == "" {
		return "", "", fmt.Errorf("invalid target: %s — expected <user@host>", target)
	}
	return user, host, nil
}
