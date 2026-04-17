package main

import (
	"github.com/chrispian/stack-explorer/internal/api"
	"github.com/spf13/cobra"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the REST API server",
	Long:  "Start an HTTP server exposing the Stack Explorer REST API for the Sigil frontend.",
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := api.NewServer(store)
		return srv.ListenAndServe(servePort)
	},
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "port to listen on")
	rootCmd.AddCommand(serveCmd)
}
