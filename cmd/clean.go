package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ndzuma/probeTool/internal/paths"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean probe reports and cache",
	Long:  `Removes all probe report markdown files and clears cache.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Clean probe markdown files
		probesDir := paths.GetProbesDir()
		pattern := filepath.Join(probesDir, "*.md")
		matches, err := filepath.Glob(pattern)
		if err == nil {
			for _, file := range matches {
				os.Remove(file)
			}
			fmt.Printf("✅ Cleaned %d probe report(s)\n", len(matches))
		}

		// Clean cache
		cacheDir := paths.GetCacheDir()
		os.RemoveAll(cacheDir)
		fmt.Println("✅ Cleared cache")
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
