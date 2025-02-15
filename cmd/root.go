package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "mcp",
		Short: "MCP (Model Context Protocol) server and client",
		Long:  `MCP (Model Context Protocol) server and client`,
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	l := log.New(logger.Writer(), "", log.LstdFlags)

	ctx := root.Context()

	root.AddCommand(NewServerCmd(ctx, l))
	root.AddCommand(NewClientCmd(ctx, l))

	return root
}
