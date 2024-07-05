package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Hu13er/lowdb"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:     "get",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"cat", "show"},
	RunE: func(cmd *cobra.Command, args []string) error {
		socketfile := Config.WithCobra(cmd).SocketFilePath()
		c := lowdb.NewKVStoreHTTPClient(socketfile)
		key := args[0]
		metadata, _ := cmd.Flags().GetBool("metadata")

		kvm, err := c.Get(key)
		if err != nil {
			return err
		}

		if metadata {
			fmt.Fprintf(os.Stderr, "key: %s\n", kvm.Key)
			fmt.Fprintf(os.Stderr, "revision: %d\n", kvm.Revision)
			for k, vs := range kvm.Headers {
				fmt.Fprintf(os.Stderr, "%s: %s\n", k, strings.Join(vs, ","))
			}
		}
		io.Copy(os.Stdout, bytes.NewReader(kvm.Value))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolP("metadata", "m", false, "show metadata in stderr")
}
