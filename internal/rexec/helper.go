package rexec

import (
	"fmt"
	"log"
	"os"
	"sshtools/internal/pkg/rssh"
	"sshtools/pkg/version"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func initConfig() {
	if cfgFile == "" {
		return
	}
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
}

func addCliFlags(flags *pflag.FlagSet) {
	flags.StringP("addrs", "a", "", "'host:port,host:port,...', The ssh server addresses, the falg is mutually exclusive with other flag '--config'")
	flags.StringP("username", "u", "root", "The ssh server username")
	flags.StringP("password", "p", "", "The ssh server password")
	flags.StringP("filename", "f", "", "A script file passed to the ssh server for execution, the flag is mutually exclusive with other flag '--cmd'")
	flags.String("cmd", "", "A command passed to the ssh server for execution, the flag is mutually exclusive with other flag '--filename'")
}

func printVersionAndExist() {
	if ver {
		info := version.New()
		fmt.Println(info)
		os.Exit(0)
	}
}

func checkArgs(cmd *cobra.Command) {
	if len(os.Args) == 1 {
		cmd.Help()
		os.Exit(0)
	}
}

func prettyPrint(resps []rssh.Response) {
	success := color.New(color.FgGreen)
	failed := color.New(color.FgRed)

	for _, resp := range resps {
		if resp.ExitStatus == 0 {
			success.Printf(">>> %s\n", resp.Addr)
			fmt.Println(resp.Output)
		} else {
			failed.Printf(">>> %s\n", resp.Addr)
			fmt.Println(resp.Err)
		}
		fmt.Println()
	}
}

func getClientConfigs() ([]rssh.ClientConfig, error) {
	cfgs := []rssh.ClientConfig{}

	addrs := viper.Get("addrs")
	v, ok := addrs.(string)
	if ok {
		l := strings.Split(v, ",")
		for _, addr := range l {
			cfg := rssh.ClientConfig{
				Addr:     addr,
				Username: viper.GetString("username"),
				Password: viper.GetString("password"),
			}
			cfgs = append(cfgs, cfg)
		}
	} else {
		viper.UnmarshalKey("addrs", &cfgs)
	}

	for i, v := range cfgs {
		addrS := strings.Split(v.Addr, ":")
		if len(addrS) == 2 {
			continue
		}
		if len(addrS) == 1 {
			defaultPort := "22"
			addr := fmt.Sprintf("%s:%s", addrS[0], defaultPort)
			v.Addr = addr
			cfgs[i] = v
			continue
		}
		return nil, fmt.Errorf("Host addr %s is incorrect", v.Addr)
	}

	return cfgs, nil
}
