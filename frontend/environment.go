package frontend

import (
	"io"
	"log"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/ast"
	"github.com/stackoverflow/novah-go/frontend/lexer"
	"github.com/stackoverflow/novah-go/frontend/parser"
	tc "github.com/stackoverflow/novah-go/frontend/typechecker"
)

const ERROR_THRESHOLD = 10

// The environment where a full compilation
// process takes place.
type Environment struct {
	opts    Options
	modules map[string]tc.FullModuleEnv
	errors  []data.CompilerProblem
}

func NewEnviroment(opts Options) *Environment {
	return &Environment{opts: opts, modules: make(map[string]tc.FullModuleEnv), errors: make([]data.CompilerProblem, 0)}
}

func (env *Environment) ParseSources(srcs []Source) (map[string]tc.FullModuleEnv, []data.CompilerProblem) {
	return env.parseSources(srcs, false)
}

func (env *Environment) parseSources(srcs []Source, isStdlib bool) (map[string]tc.FullModuleEnv, []data.CompilerProblem) {
	modMap := make(map[string]*data.DagNode[string, ast.SModule])
	modGraph := data.NewDag[string, ast.SModule](len(srcs))

	alreadySeenPath := data.NewSet[string]()
	for _, src := range srcs {
		path := src.Path
		// TODO: check for duplicate modules
		// don't parse the same path
		if alreadySeenPath.Contains(path) {
			continue
		}
		alreadySeenPath.Add(path)

		if env.opts.Verbose {
			log.Default().Printf("parsing %s\n", path)
		}

		src.WithReader(func(reader io.Reader) {
			lex := lexer.New(src.Path, reader)
			parser := parser.NewParser(lex)
			mod, errs := parser.ParseFullModule()
			env.errors = append(env.errors, errs...)

			module := mod.Name.Val
			node := data.NewDagNode(module, mod)
			if _, has := modMap[module]; has {
				env.errors = append(env.errors, duplicateError(mod, path))
			}
			modMap[module] = node
		})
	}

	if shouldStop(env.errors) {
		return nil, env.errors
	}

	if len(modMap) == 0 {
		if env.opts.Verbose {
			log.Default().Println("No files to compile")
		}
		return env.modules, nil
	}

	// add all nodes to the graph
	modGraph.AddNodes(data.MapValues(modMap)...)
	// link all the nodes
	for _, node := range modMap {
		for _, imp := range node.Data.Imports {
			if other, has := modMap[imp.Module.Val]; has {
				other.Link(node)
			}
		}
	}

	if cycle, hasCycle := modGraph.FindCycle(); hasCycle {
		env.reportCycle(cycle)
		return nil, env.errors
	}

	omods := modGraph.Toposort()
	for _, modNode := range omods.ToSlice() {
		mod := modNode.Data
		checker := tc.NewTypechecker()
		importErrs := resolveImports(&mod, env.modules, checker.Env())
		env.errors = append(env.errors, importErrs...)
		if shouldStop(env.errors) {
			return nil, env.errors
		}

		if env.opts.Verbose {
			log.Default().Printf("typechecking %s\n", mod.Name.Val)
		}

		desugar := NewDesugar(mod, checker)
		canon, err := desugar.Desugar()
		if err != nil {
			env.errors = append(env.errors, err.(data.CompilerProblem))
			env.errors = append(env.errors, desugar.errors...)
			return nil, env.errors
		}
		env.errors = append(env.errors, desugar.errors...)
		if shouldStop(env.errors) {
			return nil, env.errors
		}

		menv, err := checker.Infer(canon)
		if err != nil {
			env.errors = append(env.errors, err.(data.CompilerProblem))
			env.errors = append(env.errors, checker.Errors()...)
			return nil, env.errors
		}
		env.errors = append(env.errors, checker.Errors()...)
		if shouldStop(env.errors) {
			return nil, env.errors
		}

		env.modules[mod.Name.Val] = tc.FullModuleEnv{Env: menv, Ast: canon, TypeVarsMap: checker.TypeVarMap, Comment: mod.Comment, IsStdlib: isStdlib}
	}
	return env.modules, nil
}

// Optimize the AST and generate go code
func (env *Environment) GenerateCode(output string, dryRun bool) {

}

func duplicateError(mod ast.SModule, path string) data.CompilerProblem {
	return data.CompilerProblem{
		Msg:      data.DuplicateModule(mod.Name.Val),
		Span:     mod.Name.Span,
		Filename: path,
		Module:   mod.Name.Val,
		Severity: data.ERROR,
	}
}

func (env *Environment) reportCycle(nodes []data.DagNode[string, ast.SModule]) {
	msg := data.CycleFound(data.MapSlice(nodes, func(t data.DagNode[string, ast.SModule]) string { return t.Val }))
	for _, node := range nodes {
		mod := node.Data
		err := data.CompilerProblem{Msg: msg, Span: mod.Span, Filename: mod.SourceName, Module: mod.Name.Val, Severity: data.ERROR}
		env.errors = append(env.errors, err)
	}
}

func shouldStop(errs []data.CompilerProblem) bool {
	count := 0
	for _, err := range errs {
		if err.Severity == data.FATAL {
			return true
		}
		if err.Severity == data.ERROR {
			count++
		}
	}
	return count > ERROR_THRESHOLD
}
