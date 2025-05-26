package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version of kubectl-tks",
	Long:  `print version of kubectl-tks`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v0.8.0")
	},
}
