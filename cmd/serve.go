package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Hu13er/lowdb"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use: "serve",
	RunE: func(cmd *cobra.Command, args []string) error {
		sockfile := Config.WithCobra(cmd).SocketFilePath()
		log.Printf("Listening on %q...\n", sockfile)
		socket, err := net.Listen("unix", sockfile)
		if err != nil {
			log.Println("Error openning socket:", err)
			return err
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			os.Remove(sockfile)
			os.Exit(0)
		}()

		storage := lowdb.NewMemoryKVStore()
		handler := (&lowdb.KVStoreHTTP{
			Store: storage,
		}).Handlers()

		server := http.Server{
			Handler: handler,
		}

		if err := server.Serve(socket); err != nil {
			log.Println("Error serving http:", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
