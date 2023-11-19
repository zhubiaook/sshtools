package rcp

import (
	"fmt"
	"log"
	"sshtools/internal/pkg/rsftp"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewUploadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "upload",
		Short:        "Upload files to multiple SSH server",
		SilenceUsage: true,
		RunE:         runUpload,
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&cfgFile, "config", "c", "", "The ssh server configuration file")
	addCliFlags(flags)
	cmd.MarkPersistentFlagRequired("localpath")
	cmd.MarkPersistentFlagRequired("remotepath")
	cmd.MarkFlagsMutuallyExclusive("addrs", "config")

	return cmd
}

func runUpload(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		log.Fatal(err)
	}

	cfgs, err := getClientConfigs()
	if err != nil {
		log.Fatal(err)
	}

	mc, err := rsftp.NewMultiClient(cfgs)
	if err != nil {
		log.Fatal(err)
	}
	defer mc.Close()

	localPath := viper.GetString("localpath")
	remotePath := viper.GetString("remotepath")
	force := viper.GetBool("force")
	resps := mc.UploadFiles(localPath, remotePath, force)
	prettyPrint(resps)

	return nil
}
