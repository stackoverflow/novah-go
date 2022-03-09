package compilecmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackoverflow/novah-go/compiler"
)

var CompileCmd = &cobra.Command{
	Use:   "compile [novah sources]",
	Short: "compile novah source files to go",
	Long:  `compile novah sources to go and store it in the output folder`,
	Run:   runCompile,
}

var output string
var verbose *bool

func init() {
	CompileCmd.Flags().StringVarP(&output, "output", "o", "output", "output directoy for generated files")
	verbose = CompileCmd.Flags().BoolP("verbose", "v", false, "print more information about the compilation")
}

func runCompile(cmd *cobra.Command, args []string) {
	if *verbose {
		fmt.Printf("compiling to %s...\n", output)
	}

	sources := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.HasSuffix(arg, ".novah") {
			sources = append(sources, arg)
		}
	}

	compiler := compiler.NewCompiler(sources, compiler.Options{Verbose: *verbose})
	compiler.Run(output, false)
}
