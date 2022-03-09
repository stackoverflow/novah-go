// Generates go code
package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/stackoverflow/novah-go/compiler/ast"
	"github.com/stackoverflow/novah-go/data"
)

type Codegen struct {
	pack   ast.GoPackage
	sb     strings.Builder
	tab    string
	toInit []ast.GoVarDecl
}

func NewCodegen(pack ast.GoPackage) *Codegen {
	return &Codegen{pack: pack, toInit: make([]ast.GoVarDecl, 0, 2)}
}

func (c *Codegen) Run() string {
	c.write("//line ", c.pack.SourceName, ":", strconv.Itoa(c.pack.Pos.Line), "\n")
	c.write("package ", c.pack.Name, "\n\n")

	c.sb.WriteString(`func __if[T any](cond bool, then, els func() T) T {
  if cond {
	  return then()
  } else {
	  return els()
  }
}

`)

	for _, decl := range c.pack.Decls {
		c.genDecl(decl)
		c.sb.WriteString("\n\n")
	}
	return c.sb.String()
}

func (c *Codegen) genDecl(decl ast.GoDecl) {
	c.writePosLn(decl.GetPos())
	switch d := decl.(type) {
	case ast.GoInterface:
		c.genInterface(d)
	case ast.GoStruct:
		c.genStruct(d)
	case ast.GoConstDecl:
		{
			c.write("const ", d.Name, " = ")
			c.sb.WriteString(d.Val.V)
		}
	case ast.GoVarDecl:
		{
			c.write("var ", d.Name)
			c.genType(d.Type)
			c.toInit = append(c.toInit, d)
		}
	case ast.GoFuncDecl:
		c.genFuncDecl(d)
	default:
		panic("got unknow GoDecl in codegen")
	}
}

func (c *Codegen) genFuncDecl(d ast.GoFuncDecl) {
	c.write("func ", d.Name, "(")
	i := 0
	for par, typ := range d.Params {
		if i > 0 {
			c.sb.WriteString(", ")
		}
		c.write(par, " ")
		c.genType(typ)
		i++
	}
	c.sb.WriteRune(')')
	if len(d.Returns) > 0 {
		c.sb.WriteRune(' ')
		if len(d.Returns) == 1 {
			c.genType(d.Returns[0])
		} else {
			c.sb.WriteRune('(')
			for i, ret := range d.Returns {
				if i > 0 {
					c.sb.WriteString(", ")
				}
				c.genType(ret)
			}
			c.sb.WriteRune(')')
		}
	}
	if d.Body == nil {
		c.sb.WriteString(" {}")
	} else {
		c.sb.WriteString(" {\n")
		c.withTab(func() {
			c.genExpr(*d.Body)
		})
		c.sb.WriteString("\n}\n\n")
	}
}

func (c *Codegen) genStruct(d ast.GoStruct) {
	c.write("type ", d.Name, " struct {")
	c.withTab(func() {
		for name, field := range d.Fields {
			c.sb.WriteRune('\n')
			c.writeTab(name, " ")
			c.genType(field)
		}
	})
	c.sb.WriteString("\n}\n\n")
}

func (c *Codegen) genInterface(d ast.GoInterface) {
	c.write("type ", d.Name, " interface {")
	c.withTab(func() {
		for _, method := range d.Methods {
			c.sb.WriteRune('\n')
			c.writeTab(method.Name, "(")
			for i, arg := range method.Args {
				if i > 0 {
					c.sb.WriteString(", ")
				}
				c.genType(arg)
			}
			c.sb.WriteString(")")
			if len(method.Ret) == 1 {
				c.sb.WriteRune(' ')
				c.genType(method.Ret[0])
			}
			if len(method.Ret) > 1 {
				c.sb.WriteString(" (")
				for i, ret := range method.Ret {
					if i > 0 {
						c.sb.WriteString(", ")
					}
					c.genType(ret)
				}
				c.sb.WriteRune(')')
			}
		}
	})
	c.sb.WriteString("\n}\n\n")
}

func (c *Codegen) genExpr(exp ast.GoExpr) {
	c.writePos(exp.GetPos())
	switch e := exp.(type) {
	case ast.GoConst:
		c.sb.WriteString(e.V)
	case ast.GoVar:
		{
			if e.Package != "" {
				c.write(e.Package, ".")
			}
			c.write(e.Name)
		}
	case ast.GoFunc:
		{
			c.sb.WriteString("func (")
			i := 0
			for v, ty := range e.Args {
				if i > 0 {
					c.sb.WriteString(", ")
				}
				c.write(v, " ")
				c.genType(ty)
				i++
			}
			c.sb.WriteRune(')')
			if len(e.Returns) > 0 {
				c.sb.WriteRune(' ')
				if len(e.Returns) == 1 {
					c.genType(e.Returns[0])
				} else {
					c.sb.WriteRune('(')
					for i, ty := range e.Returns {
						if i > 0 {
							c.sb.WriteString(", ")
						}
						c.genType(ty)
					}
					c.sb.WriteRune(')')
				}
			}
			c.write(" {\n")
			c.withTab(func() {
				c.sb.WriteString(c.tab)
				c.genExpr(e.Body)
			})
			c.write("\n", c.tab, "}")
		}
	case ast.GoCall:
		{
			c.genExpr(e.Fn)
			c.sb.WriteRune('(')
			for i, arg := range e.Args {
				if i > 0 {
					c.sb.WriteString(", ")
				}
				c.genExpr(arg)
			}
			c.sb.WriteRune(')')
		}
	case ast.GoReturn:
		{
			c.sb.WriteString("return ")
			c.genExpr(e.Exp)
		}
	case ast.GoIf:
		{
			c.sb.WriteString("if ")
			c.genExpr(e.Cond)
			c.sb.WriteString(" {\n")
			c.withTab(func() {
				c.sb.WriteString(c.tab)
				c.genExpr(e.Then)
			})
			c.write("\n", c.tab, "} else {\n")
			c.withTab(func() {
				c.sb.WriteString(c.tab)
				c.genExpr(e.Else)
			})
			c.write("\n", c.tab, "}")
		}
	case ast.GoVarDef:
		{
			c.write("var ", e.Name, " ")
			c.genType(e.Type)
		}
	case ast.GoLet:
		{
			c.write(e.Binder, " := ")
			c.genExpr(e.BindExpr)
		}
	case ast.GoSetvar:
		{
			c.write(e.Name, " = ")
			c.genExpr(e.Exp)
		}
	case ast.GoStmts:
		for i, exp := range e.Exps {
			if i > 0 {
				c.write("\n", c.tab)
			}
			c.genExpr(exp)
		}
	case ast.GoUnit:
		c.sb.WriteString("nil")
	case ast.GoNil:
		c.sb.WriteString("nil")
	case ast.GoWhile:
		{
			c.sb.WriteString("for ")
			c.genExpr(e.Cond)
			c.withTab(func() {
				c.write(" {\n", c.tab)
				for i, exp := range e.Exps {
					if i > 0 {
						c.write("\n", c.tab)
					}
					c.genExpr(exp)
				}
			})
		}
	default:
		panic("unknow GoExpr in codegen")
	}
}

func (c *Codegen) genType(typ ast.GoType) {
	switch t := typ.(type) {
	case ast.GoTConst:
		{
			if t.Package != "" {
				c.write(t.Package, ".")
			}
			c.write(t.Name)
		}
	case ast.GoTFunc:
		{
			c.sb.WriteString("func(")
			c.genType(t.Arg)
			c.sb.WriteString(") ")
			c.genType(t.Ret)
		}
	}
}

func (c *Codegen) write(strs ...string) {
	for _, s := range strs {
		c.sb.WriteString(s)
	}
}

func (c *Codegen) writeTab(strs ...string) {
	c.sb.WriteString(c.tab)
	c.write(strs...)
}

func (c *Codegen) writePos(pos data.Pos) {
	c.write("/*line :", strconv.Itoa(pos.Line), ":", strconv.Itoa(pos.Col), "*/")
}

func (c *Codegen) writePosLn(pos data.Pos) {
	c.write("//line :", strconv.Itoa(pos.Line), ":", strconv.Itoa(pos.Col), "\n")
}

func (c *Codegen) withTab(f func()) {
	tmp := c.tab
	c.tab = fmt.Sprintf("%s  ", c.tab)
	f()
	c.tab = tmp
}
