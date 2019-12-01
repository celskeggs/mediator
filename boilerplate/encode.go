package main

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/predefs"
	"go/ast"
	"sort"
	"strings"
	"unicode"
)

func (t *PreparedProc) ParamNums() []int {
	var result []int
	for i := 0; i < t.ParamCount; i++ {
		result = append(result, i)
	}
	return result
}

func (t *PreparedVar) ConvertTo() []string {
	if ident, ok := t.Type.(*ast.Ident); ok {
		switch ident.Name {
		case "bool":
			return []string{"types.Bool(", ")"}
		}
	}
	return []string{"", ""}
}

func (t *PreparedVar) ConvertFrom() []string {
	if ident, ok := t.Type.(*ast.Ident); ok {
		switch ident.Name {
		case "bool":
			return []string{"types.Unbool(", ")"}
		default:
			if len(ident.Name) >= 1 && unicode.IsUpper(rune(ident.Name[0])) {
				// local custom type
				return []string{"", fmt.Sprintf(".(%s.%s)", t.PackageShort, ident.Name)}
			}
		}
	} else if sel, ok := t.Type.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			return []string{"", fmt.Sprintf(".(%s.%s)", ident.Name, sel.Sel.Name)}
		}
	}
	return []string{"", ""}
}

// in order from subclass to superclass
func (t *TypeInfo) Chunks(tree *TreeInfo) ([]*TypeInfo, error) {
	if t.Parent == "/" {
		return []*TypeInfo{t}, nil
	}
	parent, ok := tree.Paths[t.Parent]
	if !ok {
		return nil, fmt.Errorf("no such path %s as parent of %s", t.Parent, t.Path)
	}
	chunks, err := parent.Chunks(tree)
	if err != nil {
		return nil, err
	}
	return append([]*TypeInfo{t}, chunks...), nil
}

func (t *TypeInfo) EncodedChunks(tree *TreeInfo) ([]PreparedChunk, error) {
	chunks, err := t.Chunks(tree)
	if err != nil {
		return nil, err
	}
	var encodedChunks []PreparedChunk
	for _, chunk := range chunks {
		enc, err := chunk.Encode(tree)
		if err != nil {
			return nil, err
		}
		encodedChunks = append(encodedChunks, enc)
	}
	return encodedChunks, nil
}

func ImportName(imp *ast.ImportSpec) string {
	if imp.Name != nil {
		return imp.Name.Name
	}
	parts := strings.Split(Unquote(imp.Path.Value), "/")
	return parts[len(parts)-1]
}

func (t *TypeInfo) Imports(tree *TreeInfo) ([]string, error) {
	chunks, err := t.Chunks(tree)
	if err != nil {
		return nil, err
	}

	imports := map[string]struct{}{
		"github.com/celskeggs/mediator/platform/types": {},
	}
	for _, chunk := range chunks {
		imports[chunk.Package] = struct{}{}
		for _, vi := range chunk.Vars {
			if sel, ok := vi.Type.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					var found string
					for _, imp := range vi.DefiningImports {
						if ImportName(imp) == ident.Name {
							found = Unquote(imp.Path.Value)
						}
					}
					if found == "" {
						return nil, fmt.Errorf("could not find import %s", ident.Name)
					}
					imports[found] = struct{}{}
				}
			}
		}
	}
	var importList []string
	for i, _ := range imports {
		importList = append(importList, i)
	}
	sort.Strings(importList)
	return importList, nil
}

func (t *TypeInfo) AllVars(tree *TreeInfo) ([]PreparedGetter, []PreparedVar, []PreparedProc, error) {
	chunks, err := t.Chunks(tree)
	if err != nil {
		return nil, nil, nil, err
	}
	var getters []PreparedGetter
	var vars []PreparedVar
	var procs []PreparedProc
	hasgetter := map[string]bool{}
	hasproc := map[string]bool{}
	// order of chunks is such that subclasses override superclass getters and procs
	for _, chunk := range chunks {
		for _, vi := range chunk.Vars {
			vars = append(vars, PreparedVar{
				VarInfo:      vi,
				StructName:   chunk.StructName,
				PackageShort: chunk.PackageShort(),
			})
		}
		for _, gi := range chunk.Getters {
			if hasgetter[gi.FieldName] {
				continue
			}
			hasgetter[gi.FieldName] = true
			if !gi.HasGetter {
				return nil, nil, nil, fmt.Errorf("no getter for field: %s", gi.FieldName)
			}
			getters = append(getters, PreparedGetter{
				GetterInfo: *gi,
				StructName: chunk.StructName,
			})
		}
		for _, pi := range chunk.Procs {
			if hasproc[pi.Name] {
				continue
			}
			hasproc[pi.Name] = true
			procs = append(procs, PreparedProc{
				ProcInfo:   pi,
				StructName: chunk.StructName,
			})
		}
	}
	sort.Slice(vars, func(i, j int) bool {
		return vars[i].FieldName < vars[j].FieldName
	})
	sort.Slice(getters, func(i, j int) bool {
		return getters[i].FieldName < getters[j].FieldName
	})
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].Name < procs[j].Name
	})
	return getters, vars, procs, nil
}

func (t *TypeInfo) PackageShort() string {
	parts := strings.Split(t.Package, "/")
	return parts[len(parts)-1]
}

func (t *TypeInfo) Encode(tree *TreeInfo) (PreparedChunk, error) {
	if !t.FoundConstructor {
		return PreparedChunk{}, fmt.Errorf("no constructor for %s", t.Path)
	}
	return PreparedChunk{
		PackageShort: t.PackageShort(),
		Package:      t.Package,
		StructName:   t.StructName,
		Vars:         t.Vars,
	}, nil
}

func (t *TreeInfo) Encode() ([]*PreparedImplementation, error) {
	var pis []*PreparedImplementation
	for _, typeInfo := range t.Paths {
		chunks, err := typeInfo.EncodedChunks(t)
		if err != nil {
			return nil, err
		}
		imports, err := typeInfo.Imports(t)
		if err != nil {
			return nil, err
		}
		parent := typeInfo.Parent
		if parent == "/" {
			parent = ""
		}
		getters, vars, procs, err := typeInfo.AllVars(t)
		if err != nil {
			return nil, err
		}
		pis = append(pis, &PreparedImplementation{
			TypePath:   typeInfo.Path,
			Type:       typeInfo.Type(),
			Imports:    imports,
			Chunks:     chunks,
			ParentPath: parent,
			Vars:       vars,
			Procs:      procs,
			Getters:    getters,
		})
	}
	return pis, nil
}

func (i *TypeInfo) Type() string {
	if len(i.Path) < 2 || i.Path[0] != '/' {
		panic("invalid type path")
	}
	var parts []string
	for _, part := range strings.Split(i.Path[1:], "/") {
		parts = append(parts, predefs.ToTitle(part))
	}
	return strings.Join(parts, "")
}
