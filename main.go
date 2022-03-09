package main

import (
	"github.com/spf13/cobra"
	compile "github.com/stackoverflow/novah-go/cmd/compile_cmd"
)

func main() {
	rootCmd := &cobra.Command{Use: "novah", Version: "0.1"}
	rootCmd.AddCommand(compile.CompileCmd)
	rootCmd.Execute()
}
