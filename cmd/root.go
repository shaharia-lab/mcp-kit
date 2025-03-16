package cmd

import (
	"github.com/spf13/cobra"
)

var configFile string

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "mcp",
		Short: "MCP (Model Context Protocol) server and client",
		Long:  `MCP (Model Context Protocol) server and client`,
	}

	root.PersistentFlags().StringVar(&configFile, "config", "config.yaml", "path to config file")

	root.AddCommand(NewServerCmd())
	root.AddCommand(NewTaskCmd())
	root.AddCommand(NewAPICmd())
	root.AddCommand(NewDevTestCmd())

	return root
}
