package cmd

import (
	"fmt"
	"os"

	"github.com/ndzuma/probeTool/internal/version"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display detailed version information about probeTool including build details and platform.`,
	Run: func(cmd *cobra.Command, args []string) {
		info := version.GetInfo()

		if jsonOutput {
			output, err := info.JSON()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(output)
		} else {
			fmt.Println(info.Detailed())
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output version information as JSON")
}
