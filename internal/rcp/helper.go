package rcp

import (
	"fmt"
	"log"
	"os"
	"sshtools/internal/pkg/rsftp"
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
	flags.StringP("addrs", "a", "", "'host:port,host:port,...', The ssh server addresses")
	flags.StringP("username", "u", "root", "The ssh server username")
	flags.StringP("password", "p", "", "The ssh server password")
	flags.StringP("localpath", "l", "", "Local file or directory")
	flags.StringP("remotepath", "r", "", "Remote file or directory")
	flags.Bool("force", false, "Force overwriting of files that already exist")
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

func prettyPrint(resps []rsftp.Response) {
	success := color.New(color.FgGreen)
	failed := color.New(color.FgRed)

	for _, resp := range resps {
		if resp.Err != nil {
			failed.Printf(">>> %s\n", resp.Addr)
			fmt.Printf("Error: %s\n", resp.Err)
		} else {
			success.Printf(">>> %s\n", resp.Addr)
			fmt.Printf("Output: %s\n", resp.Output)
		}
		fmt.Println()
	}
}

func getClientConfigs() ([]rsftp.ClientConfig, error) {
	cfgs := []rsftp.ClientConfig{}

	addrs := viper.Get("addrs")
	v, ok := addrs.(string)
	if ok {
		l := strings.Split(v, ",")
		for _, addr := range l {
			cfg := rsftp.ClientConfig{
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
