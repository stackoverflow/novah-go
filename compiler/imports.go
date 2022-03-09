package compiler

import (
	"fmt"

	"github.com/stackoverflow/novah-go/compiler/ast"
	tc "github.com/stackoverflow/novah-go/compiler/typechecker"
	"github.com/stackoverflow/novah-go/data"
)

func resolveImports(mod *ast.SModule, mods map[string]tc.FullModuleEnv, env *tc.Env) []data.CompilerProblem {

	makeError := func(span data.Span, sev data.Severity) func(string) data.CompilerProblem {
		return func(msg string) data.CompilerProblem {
			return data.CompilerProblem{Msg: msg, Span: span, Filename: mod.SourceName, Module: mod.Name.Val, Severity: sev}
		}
	}

	resolved := make(map[string]string)
	resolvedTypealias := make([]ast.STypeAliasDecl, 0)
	errors := make([]data.CompilerProblem, 0)
	for _, imp := range mod.Imports {
		mkError := makeError(imp.Span, data.ERROR)
		mname := imp.Module.Val
		smod, found := mods[mname]
		if !found {
			errors = append(errors, mkError(data.ModuleNotFound(mname)))
			continue
		}
		m := smod.Env
		typealiases := getTypealiases(mname, mods)

		// if there's an alias, add the whole module aliased
		if imp.Alias != "" {
			alias := fmt.Sprintf("%s.", imp.Alias)
			for name, ty := range m.Types {
				if ty.Visibility == ast.PUBLIC {
					resolved[fmt.Sprintf("%s%s", alias, name)] = mname
					env.ExtendType(fmt.Sprintf("%s.%s", mname, name), ty.Type)
				}
			}
			for name, d := range m.Decls {
				resolved[fmt.Sprintf("%s%s", alias, name)] = mname
				env.Extend(fmt.Sprintf("%s.%s", mname, name), d.Type)
				if d.IsInstance {
					env.ExtendInstance(fmt.Sprintf("%s.%s", mname, name), d.Type, false)
				}
			}
			for _, ta := range typealiases {
				if ta.Visibility == ast.PUBLIC {
					resolvedTypealias = append(resolvedTypealias, ta)
				}
			}
		}

		for _, ref := range imp.Defs {
			refname := ref.Name.Val
			if ref.Tag == ast.VAR {
				declRef, found := m.Decls[refname]
				if !found {
					errors = append(errors, mkError(data.CannotFindInModule(fmt.Sprintf("declaration %s", refname), mname)))
					continue
				}
				if declRef.Visibility == ast.PRIVATE {
					errors = append(errors, mkError(data.CannotImportInModule(fmt.Sprintf("declaration %s", refname), mname)))
					continue
				}
				fname := fmt.Sprintf("%s.%s", mname, refname)
				resolved[refname] = mname
				env.Extend(fname, declRef.Type)
				if declRef.IsInstance {
					env.ExtendInstance(fname, declRef.Type, false)
				}
			} else {
				talias, found := typealiases[refname]
				if found {
					if talias.Visibility == ast.PRIVATE {
						errors = append(errors, mkError(data.CannotImportInModule(fmt.Sprintf("type %s", refname), mname)))
						continue
					}
					resolvedTypealias = append(resolvedTypealias, talias)
					continue
				}
				declRef, found := m.Types[refname]
				if !found {
					errors = append(errors, mkError(data.CannotFindInModule(fmt.Sprintf("type %s", ref.Name.Span), mname)))
					continue
				}
				if declRef.Visibility == ast.PRIVATE {
					errors = append(errors, mkError(data.CannotImportInModule(fmt.Sprintf("type %s", refname), mname)))
					continue
				}
				env.ExtendType(fmt.Sprintf("%s.%s", mname, refname), declRef.Type)
				resolved[refname] = mname
				if ref.All {
					// import all constructors
					for _, ctor := range declRef.Ctors {
						ctorDecl := m.Decls[ctor] // cannot fail
						if ctorDecl.Visibility == ast.PRIVATE {
							errors = append(errors, mkError(data.CannotImportInModule(fmt.Sprintf("constructor %s", ctor), mname)))
							continue
						}
						env.Extend(fmt.Sprintf("%s.%s", mname, ctor), ctorDecl.Type)
						resolved[ctor] = mname
					}
				} else {
					// import defined constructors
					for _, ctor := range ref.Ctors {
						name := ctor.Val
						ctorDecl, found := m.Decls[name]
						if !found {
							errors = append(errors, mkError(data.CannotFindInModule(fmt.Sprintf("constructor %s", name), mname)))
							continue
						}
						if ctorDecl.Visibility == ast.PRIVATE {
							errors = append(errors, mkError(data.CannotImportInModule(fmt.Sprintf("constructor %s", name), mname)))
							continue
						}
						env.Extend(fmt.Sprintf("%s.%s", mname, name), ctorDecl.Type)
						resolved[name] = mname
					}
				}
			}
		}
	}
	mod.ResolvedImports = resolved
	mod.ResolvedTypealiases = resolvedTypealias
	return errors
}

func getTypealiases(name string, mods map[string]tc.FullModuleEnv) map[string]ast.STypeAliasDecl {
	mod, found := mods[name]
	m := make(map[string]ast.STypeAliasDecl)
	if !found {
		return m
	}
	aliases := mod.Aliases
	if aliases == nil || len(aliases) == 0 {
		return m
	}
	for _, alias := range aliases {
		m[alias.Name] = alias
	}
	return m
}
