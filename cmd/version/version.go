package version

import (
	"fmt"
	"github.com/buexplain/go-watch/cmd"
	"github.com/spf13/cobra"
)

var Version *cobra.Command

func init() {
	Version = &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("go-watch v0.0.0")
		},
	}
	cmd.Root.AddCommand(Version)
}
