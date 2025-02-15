package main

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/mcp-kit/cmd"
	"os"
)

func main() {
	ctx := context.Background()
	rootCmd := cmd.NewRootCmd()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
