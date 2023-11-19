package rcp

import (
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	ver     bool
)

func NewRCopyCommand() *cobra.Command {
	cobra.OnInitialize(initConfig)
	cmd := &cobra.Command{
		Use:          "rcp",
		Short:        "Copy files form/to multiple SSH server",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			printVersionAndExist()
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&ver, "version", "V", false, "Print version and exist")

	cmd.AddCommand(NewUploadCommand())
	cmd.AddCommand(NewDownloadCommand())

	checkArgs(cmd)
	return cmd
}
