package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	Config _Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "lowdb",
}

func init() {
	rootCmd.PersistentFlags().StringP("unix-socket", "U", "", "unix-socket file path (default is /tmp/lowdb.sock)")
}

type _Config struct{}

func (_Config) WithCobra(cmd *cobra.Command) _ConfigCtx {
	return _ConfigCtx{cobraCmd: cmd}
}

type _ConfigCtx struct {
	cobraCmd *cobra.Command
}

func (c _ConfigCtx) SocketFilePath() string {
	if env := os.Getenv("LOWDB_SOCK"); env != "" {
		return env
	}
	if flag, _ := c.cobraCmd.PersistentFlags().GetString("unix-socket"); flag != "" {
		return flag
	}
	// default
	return "/tmp/lowdb.sock"
}
