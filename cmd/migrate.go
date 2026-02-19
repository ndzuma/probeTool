package cmd

import (
	"fmt"
	"os"

	"github.com/ndzuma/probeTool/internal/paths"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate config from old location to OS-standard paths",
	Long:  `Migrates configuration and data from ~/.probe to OS-appropriate locations.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !paths.NeedsMigration() {
			fmt.Println("✅ No migration needed - already using new paths or no old data found")
			return
		}

		if err := paths.Migrate(); err != nil {
			fmt.Printf("❌ Migration failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
