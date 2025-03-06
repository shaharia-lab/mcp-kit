package main

import (
	"context"
	"fmt"
	"os"

	"github.com/shaharia-lab/mcp-kit/cmd"
)

func main() {
	ctx := context.Background()
	rootCmd := cmd.NewRootCmd()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
