package main

import (
	"log"

	"github.com/Hu13er/lowdb"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"del", "remove", "rm"},
	RunE: func(cmd *cobra.Command, args []string) error {
		socketfile := Config.WithCobra(cmd).SocketFilePath()
		c := lowdb.NewKVStoreHTTPClient(socketfile)
		key := args[0]
		rev, err := cmd.Flags().GetInt("revision")
		if err != nil {
			log.Fatalln(err)
		}

		kvm := lowdb.KeyValueMetadata{
			Key:      key,
			Revision: rev,
		}
		return c.Delete(kvm)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().IntP("revision", "r", -1, "delete only if revisions are same. (use for lock mechanism)")
}
