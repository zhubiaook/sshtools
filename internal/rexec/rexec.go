package rexec

import (
	"fmt"
	"log"
	"sshtools/internal/pkg/rssh"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	ver     bool
)

func NewRExecCommand() *cobra.Command {
	cobra.OnInitialize(printVersionAndExist, initConfig)
	cmd := &cobra.Command{
		Use:          "rexec",
		Short:        "Execute command or script concurrently on multiple SSH servers",
		SilenceUsage: true,
		RunE:         run,
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
	}

	flags := cmd.PersistentFlags()
	flags.StringVarP(&cfgFile, "config", "c", "", "The ssh server configuration file, the flag is mutually exclusive with other flag '--addrs'")
	addCliFlags(flags)
	flags.BoolVarP(&ver, "version", "V", false, "Print version information and exist")
	cmd.MarkFlagsMutuallyExclusive("config", "addrs")
	cmd.MarkFlagsMutuallyExclusive("cmd", "filename")

	checkArgs(cmd)
	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		log.Fatal(err)
	}

	cfgs, err := getClientConfigs()
	if err != nil {
		log.Fatal(err)
	}

	mc, err := rssh.NewMultiClient(cfgs)
	if err != nil {
		log.Fatal(err)
	}
	defer mc.Close()

	command := viper.GetString("cmd")
	if command != "" {
		resps := mc.ExecCmd(command)
		prettyPrint(resps)
	}

	scriptFile := viper.GetString("filename")
	if scriptFile != "" {
		resps := mc.ExecShellScript(scriptFile)
		prettyPrint(resps)
	}

	return nil
}
