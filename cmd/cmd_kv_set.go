package main

import (
	"io"
	"log"
	"os"

	"github.com/Hu13er/lowdb"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:  "set",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		socketfile := Config.WithCobra(cmd).SocketFilePath()
		c := lowdb.NewKVStoreHTTPClient(socketfile)
		key := args[0]
		val := args[1]
		rev, err := cmd.Flags().GetInt("revision")
		if err != nil {
			log.Fatalln(err)
		}

		var buffer []byte
		if val == "-" {
			buffer, err = io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
		} else {
			buffer = []byte(val)
		}

		kvm := lowdb.KeyValueMetadata{
			Key:      key,
			Revision: rev,
			// TODO: read headers from cli
			Headers: make(map[string][]string),
			Value:   buffer,
		}
		return c.Set(kvm)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.Flags().IntP("revision", "r", -1, "write only if revisions are same. (use for lock mechanism)")
}
