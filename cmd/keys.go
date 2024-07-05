package main

import (
	"fmt"
	"strings"

	"github.com/Hu13er/lowdb"

	"github.com/spf13/cobra"
)

// keysCmd represents the keys command
var keysCmd = &cobra.Command{
	Use:     "keys",
	Aliases: []string{"key", "k"},
	RunE: func(cmd *cobra.Command, args []string) error {
		socketfile := Config.WithCobra(cmd).SocketFilePath()
		c := lowdb.NewKVStoreHTTPClient(socketfile)
		keys, err := c.Keys()
		if err != nil {
			return err
		}
		fmt.Println(strings.Join(keys, "\n"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(keysCmd)
}
