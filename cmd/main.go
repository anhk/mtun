package main

import (
	"github.com/anhk/mtun/pkg/grpc"
	"github.com/anhk/mtun/pkg/log"
	"github.com/anhk/mtun/pkg/tun"
	"github.com/spf13/cobra"
)

var (
	clientOpt grpc.ClientOption
	serverOpt grpc.ServerOption
	tunOpt    tun.TunOption
)

var clientCmd = cobra.Command{
	Use:     "client",
	Short:   "use client mode",
	Aliases: []string{"cli", "c"},
	PreRun:  checkIsRoot,
	Run: func(cmd *cobra.Command, args []string) {
		tun := tun.AllocTun()
		log.Info("Tun: %v", tun.Name)
		select {}
	},
}

var serverCmd = cobra.Command{
	Use:     "server",
	Short:   "use server mode",
	Aliases: []string{"srv", "s"},
	PreRun:  checkIsRoot,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var rootCmd = cobra.Command{
	Short: "Mtun (tunnel) for HomeMesh network",
}

func main() {
	clientCmd.PersistentFlags().StringArrayVarP(&tunOpt.Cidr, "cidr", "c", []string{}, "cidr to claim")
	clientCmd.PersistentFlags().StringVarP(&clientOpt.Token, "token", "t", "", "token to authenticate")
	clientCmd.PersistentFlags().StringVarP(&clientOpt.ServerAddr, "server", "s", "127.0.0.1", "the address of server")
	clientCmd.PersistentFlags().Uint16VarP(&clientOpt.ServerPort, "port", "p", 50051, "the port of server")

	serverCmd.PersistentFlags().StringArrayVarP(&tunOpt.Cidr, "cidr", "c", []string{}, "cidr to claim")
	serverCmd.PersistentFlags().StringVarP(&serverOpt.Token, "token", "t", "", "token to authenticate")
	serverCmd.PersistentFlags().StringVarP(&serverOpt.BindAddr, "bind", "b", "127.0.0.1", "the address to bind")
	serverCmd.PersistentFlags().Uint16VarP(&serverOpt.BindPort, "port", "p", 50051, "the port to bind")

	rootCmd.AddCommand(&clientCmd)
	rootCmd.AddCommand(&serverCmd)
	_ = rootCmd.Execute()
}
