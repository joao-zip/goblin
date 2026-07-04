package version

import "fmt"

const Version = "0.1.0"

// Banner returns the Goblin ASCII art banner with version info.
func Banner() string {
	return fmt.Sprintf("Goblin v%s — Mutation Testing for Go\n", Version)
}
