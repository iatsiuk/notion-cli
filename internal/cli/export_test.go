package cli

import (
	"github.com/spf13/cobra"

	"notion-cli/internal/config"
)

// ExecuteRootForTest executes cmd and returns the loaded config.
func ExecuteRootForTest(cmd *cobra.Command) (*config.Config, error) {
	cfg = nil
	err := cmd.Execute()
	return cfg, err
}
