package frontend

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/typechecker"
)

type Options struct {
	Verbose bool
	DevMode bool
	Stdlib  bool
}

type Compiler struct {
	sources []Source
	opts    Options
	env     *Environment
}

func NewCompiler(sources []string, opts Options) *Compiler {
	entries := data.MapSlice(sources, func(path string) Source { return Source{Path: path} })
	return &Compiler{sources: entries, opts: opts, env: NewEnviroment(opts)}
}

func (c *Compiler) Compile() (map[string]typechecker.FullModuleEnv, []data.CompilerProblem) {
	return c.env.ParseSources(c.sources)
}

func (c *Compiler) Run(output string, dryRun bool) []data.CompilerProblem {
	_, errs := c.env.ParseSources(c.sources)
	if errs != nil {
		return errs
	}
	c.env.GenerateCode(output, dryRun)
	return c.env.errors
}

func (c *Compiler) Modules() map[string]typechecker.FullModuleEnv {
	return c.env.modules
}

func (c *Compiler) Errors() []data.CompilerProblem {
	return c.env.errors
}

type Source struct {
	Path string
	Str  string
}

func (s Source) WithReader(action func(io.Reader)) {
	var reader io.Reader
	if s.Str != "" {
		reader = strings.NewReader(s.Str)
	} else {
		f, err := os.Open(s.Path)
		if err != nil {
			panic("Could not open file " + s.Path + ": " + err.Error())
		}
		defer f.Close()
		reader = bufio.NewReader(f)
	}
	action(reader)
}
