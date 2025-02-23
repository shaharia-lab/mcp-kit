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

	root.AddCommand(NewServerCmd(l))
	root.AddCommand(NewTaskCmd(l))
	root.AddCommand(NewAPICmd(l))

	return root
}
