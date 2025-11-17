package openapiproject

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"strings"
	"unicode"

	"github.com/baudii/ada-ai/internal/core/project"
	"golang.org/x/tools/go/packages"
)

const SERVER_INTERFACE = "ServerInterface"

type MethodInfo struct {
	Name   string
	Params []ParamInfo
}

type FieldInfo struct {
	Name string
	Type string
	Tags reflect.StructTag
	Doc  string
}

type ParamInfo struct {
	Name       string
	MethodName string
	Type       string
	IsStruct   bool
	StructName string
	Fields     []FieldInfo
}

func PascalToKebab(s string) string {
	var out []rune

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i != 0 {
				out = append(out, '_')
			}
			out = append(out, unicode.ToLower(r))
		} else {
			out = append(out, r)
		}
	}

	return string(out)
}

func AnalyzeMethodParams(method *types.Func, pkg *packages.Package) []ParamInfo {
	sig := method.Type().(*types.Signature)

	params := []ParamInfo{}

	typeSpecs := map[string]*ast.TypeSpec{}
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					typeSpecs[ts.Name.Name] = ts
				}
			}
		}
	}

	for i := 0; i < sig.Params().Len(); i++ {
		pv := sig.Params().At(i)
		param := ParamInfo{
			Name: pv.Name(),
			Type: pv.Type().String(),
		}

		if named, ok := pv.Type().(*types.Named); ok {
			if st, ok := named.Underlying().(*types.Struct); ok {
				param.IsStruct = true
				param.StructName = named.Obj().Name()

				if ts, ok := typeSpecs[named.Obj().Name()]; ok {
					if structNode, ok := ts.Type.(*ast.StructType); ok {
						for idx, field := range structNode.Fields.List {
							fi := FieldInfo{
								Name: field.Names[0].Name,
								Type: st.Field(idx).Type().String(),
							}

							if field.Tag != nil {
								tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
								fi.Tags = tag
							}

							if field.Doc != nil {
								for _, c := range field.Doc.List {
									fi.Doc += c.Text + "\n"
								}
							}

							param.Fields = append(param.Fields, fi)
						}
					}
				}
			}
		}

		params = append(params, param)
	}

	return params
}

// ExtractServerInterface reads the generated server code and extracts method information
// from the ServerInterface.
func ExtractServerInterface(
	pkgs []*packages.Package,
	callback func(MethodInfo, *ast.CommentGroup, project.FileHandler),
	hfile project.FileHandler,
) error {
	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()

		obj := scope.Lookup(SERVER_INTERFACE)
		if obj == nil {
			continue
		}

		if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
			for method := range iface.Methods() {
				for _, file := range pkg.Syntax {
					ast.Inspect(
						file,
						func(n ast.Node) bool {
							td, ok := n.(*ast.TypeSpec)
							if !ok || td.Name.Name != SERVER_INTERFACE {
								return true
							}

							if ifaceType, ok := td.Type.(*ast.InterfaceType); ok {
								for _, f := range ifaceType.Methods.List {
									if len(f.Names) > 0 && f.Names[0].Name == method.Name() {
										methodInfo := MethodInfo{
											Name:   method.Name(),
											Params: AnalyzeMethodParams(method, pkg),
										}
										callback(methodInfo, f.Doc, hfile)
									}
								}
							}
							return false
						})
				}
			}
		}
	}

	return nil
}

func WriteComments(out *strings.Builder, cg *ast.CommentGroup, methodName string, params []ParamInfo) {
	out.WriteString("// This file is auto-generated.\n")
	out.WriteString(fmt.Sprintf("//\n// Handler for %s method\n", methodName))

	if cg != nil {
		out.WriteString("//\n// DESCRIPTION:\n")
		for _, c := range cg.List {
			out.WriteString(c.Text + "\n")
		}
	}

	out.WriteString("//\n// PARAMETERS:\n")
	for _, prm := range params {
		out.WriteString(fmt.Sprintf("//   %s %s\n", prm.Name, prm.Type))
	}

	for _, prm := range params {
		if prm.IsStruct {
			out.WriteString(fmt.Sprintf("//\n// STRUCT %s:\n", prm.StructName))
			for _, f := range prm.Fields {
				out.WriteString(fmt.Sprintf("//   FIELD: %s (%s)\n", f.Name, f.Type))

				if tag := f.Tags.Get("json"); tag != "" {
					out.WriteString(fmt.Sprintf("//     TAG json: %s\n", tag))
				}
				if tag := f.Tags.Get("form"); tag != "" {
					out.WriteString(fmt.Sprintf("//     TAG form: %s\n", tag))
				}

				if f.Doc != "" {
					out.WriteString("//     DOC: " + f.Doc)
				}
			}
		}
	}
}
