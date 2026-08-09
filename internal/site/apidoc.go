package site

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// APIItem is one exported declaration rendered on the API page.
type APIItem struct {
	Name      string
	Anchor    string
	Kind      string
	Signature string
	Doc       string
}

// APIDocumentation is the exported API extracted from the root Go package.
type APIDocumentation struct {
	PackageName string
	ImportPath  string
	Doc         string
	Items       []APIItem
}

// ExtractAPI parses the root package and returns documentation for exported declarations only.
func ExtractAPI(projectRoot, importPath string) (APIDocumentation, error) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, projectRoot, func(info os.FileInfo) bool {
		name := info.Name()
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return APIDocumentation{}, fmt.Errorf("parse public package: %w", err)
	}
	if len(packages) != 1 {
		return APIDocumentation{}, fmt.Errorf("expected one public root package, found %d", len(packages))
	}
	var astPackage *ast.Package
	for _, parsed := range packages {
		astPackage = parsed
	}
	documented := doc.New(astPackage, importPath, 0)
	result := APIDocumentation{PackageName: documented.Name, ImportPath: importPath, Doc: documented.Doc}

	for _, value := range documented.Consts {
		result.Items = append(result.Items, valueItems(fset, value, "Constant")...)
	}
	for _, value := range documented.Vars {
		kind := "Variable"
		for _, name := range value.Names {
			if strings.HasPrefix(name, "Err") {
				kind = "Sentinel error"
			}
		}
		result.Items = append(result.Items, valueItems(fset, value, kind)...)
	}
	for _, typeDoc := range documented.Types {
		for _, value := range typeDoc.Consts {
			result.Items = append(result.Items, valueItems(fset, value, "Constant")...)
		}
		for _, value := range typeDoc.Vars {
			result.Items = append(result.Items, valueItems(fset, value, "Variable")...)
		}
		kind := "Type"
		if spec, ok := typeDoc.Decl.Specs[0].(*ast.TypeSpec); ok {
			if _, ok := spec.Type.(*ast.InterfaceType); ok {
				kind = "Interface"
			}
		}
		result.Items = append(result.Items, APIItem{
			Name: typeDoc.Name, Anchor: anchorFor(typeDoc.Name), Kind: kind,
			Signature: formatNode(fset, typeDoc.Decl), Doc: typeDoc.Doc,
		})
		for _, function := range typeDoc.Funcs {
			kind := "Function"
			if strings.HasPrefix(function.Name, "New") {
				kind = "Constructor"
			}
			result.Items = append(result.Items, functionItem(fset, function, kind))
		}
		for _, method := range typeDoc.Methods {
			result.Items = append(result.Items, functionItem(fset, method, "Method"))
		}
	}
	for _, function := range documented.Funcs {
		result.Items = append(result.Items, functionItem(fset, function, "Function"))
	}

	sort.SliceStable(result.Items, func(i, j int) bool {
		if result.Items[i].Kind == result.Items[j].Kind {
			return result.Items[i].Name < result.Items[j].Name
		}
		return result.Items[i].Kind < result.Items[j].Kind
	})
	if len(result.Items) == 0 {
		return APIDocumentation{}, fmt.Errorf("public package %s has no exported declarations", filepath.Base(projectRoot))
	}
	return result, nil
}

func valueItems(fset *token.FileSet, value *doc.Value, kind string) []APIItem {
	items := make([]APIItem, 0, len(value.Names))
	for _, name := range value.Names {
		itemKind := kind
		if strings.HasPrefix(name, "Err") {
			itemKind = "Sentinel error"
		}
		items = append(items, APIItem{Name: name, Anchor: anchorFor(name), Kind: itemKind, Signature: formatNode(fset, value.Decl), Doc: value.Doc})
	}
	return items
}

func functionItem(fset *token.FileSet, function *doc.Func, kind string) APIItem {
	return APIItem{Name: function.Name, Anchor: anchorFor(function.Name), Kind: kind, Signature: formatNode(fset, function.Decl), Doc: function.Doc}
}

func formatNode(fset *token.FileSet, node any) string {
	var output bytes.Buffer
	if err := format.Node(&output, fset, node); err != nil {
		return "declaration unavailable"
	}
	return strings.TrimSpace(output.String())
}

func anchorFor(name string) string {
	return "api-" + strings.ToLower(strings.NewReplacer(" ", "-", ".", "-").Replace(name))
}
