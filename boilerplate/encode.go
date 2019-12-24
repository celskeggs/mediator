package main

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/predefs"
	"github.com/celskeggs/mediator/util"
	"go/ast"
	"sort"
	"strings"
	"unicode"
)

func (t *ProcInfo) ParamNums() []int {
	var result []int
	for i := 0; i < t.ParamCount; i++ {
		result = append(result, i)
	}
	return result
}

func (i *ProcInfo) ProcName() string {
	name := i.Name
	if name == "<<" {
		return "OperatorWrite"
	}
	return "Proc" + name
}

func (t *PreparedVar) ConvertTo() []string {
	if ident, ok := t.Type.(*ast.Ident); ok {
		switch ident.Name {
		case "bool":
			return []string{"types.Bool(", ")"}
		case "string":
			return []string{"types.String(", ")"}
		case "int":
			return []string{"types.Int(", ")"}
		case "uint":
			return []string{"types.Int(", ")"}
		}
	}
	if at, ok := t.Type.(*ast.ArrayType); ok && at.Len == nil {
		return []string{"datum.NewListFromSlice(", ")"}
	}
	if ref, ok := getPtrTypeRef(t.Type, t.PackageShort); ok {
		if ref == "*types.Ref" {
			return []string{"", ".Dereference()"}
		}
	}
	return []string{"", ""}
}

func getTypeRef(t ast.Expr, packageShort string) (ref string, ok bool) {
	if ident, ok := t.(*ast.Ident); ok {
		if len(ident.Name) >= 1 && unicode.IsUpper(rune(ident.Name[0])) {
			// local custom type
			return fmt.Sprintf("%s.%s", packageShort, ident.Name), true
		}
	} else if sel, ok := t.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			return fmt.Sprintf("%s.%s", ident.Name, sel.Sel.Name), true
		}
	}
	return "", false
}

func getPtrTypeRef(t ast.Expr, packageShort string) (ref string, ok bool) {
	if star, ok := t.(*ast.StarExpr); ok {
		inner, ok := getTypeRef(star.X, packageShort)
		if ok {
			return "*" + inner, true
		}
	}
	return getTypeRef(t, packageShort)
}

func (t *PreparedVar) ConvertFrom() []string {
	if ref, ok := getPtrTypeRef(t.Type, t.PackageShort); ok {
		if ref == "*types.Ref" {
			return []string{"types.Reference(", ")"}
		}
		return []string{"", ".(" + ref + ")"}
	} else if ident, ok := t.Type.(*ast.Ident); ok {
		switch ident.Name {
		case "bool":
			return []string{"types.Unbool(", ")"}
		case "string":
			return []string{"types.Unstring(", ")"}
		case "int":
			return []string{"types.Unint(", ")"}
		case "uint":
			return []string{"types.Unuint(", ")"}
		}
	} else if at, ok := t.Type.(*ast.ArrayType); ok && at.Len == nil {
		if ref, ok := getTypeRef(at.Elt, t.PackageShort); ok {
			return []string{
				fmt.Sprintf("datum.ElementsAsType([]%s{}, ", ref),
				fmt.Sprintf(").([]%s)", ref),
			}
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

func (t *TypeInfo) EncodedChunks(tree *TreeInfo) ([]*PreparedChunk, error) {
	chunks, err := t.Chunks(tree)
	if err != nil {
		return nil, err
	}
	var encodedChunks []*PreparedChunk
	encodeMap := map[*SourceInfo]*PreparedChunk{}
	for _, chunk := range chunks {
		for _, source := range chunk.Sources {
			enc, err := source.Encode(tree, chunk)
			if err != nil {
				return nil, err
			}
			for _, existing := range encodedChunks {
				if existing.StructName == enc.StructName {
					util.FIXME("rename structs to support reused names better")
					return nil, fmt.Errorf("structure name reused: %s", enc.StructName)
				}
			}
			encodedChunks = append(encodedChunks, enc)
			encodeMap[source] = enc
		}
	}
	lastSeenProc := map[string]PreparedProc{}
	for i := len(chunks) - 1; i >= 0; i-- {
		chunk := chunks[i]
		for j := len(chunk.Sources) - 1; j >= 0; j-- {
			source := chunk.Sources[j]
			var preparedSuperProcs []PreparedSuperProc
			for _, proc := range source.Procs {
				last, found := lastSeenProc[proc.Name]
				if found {
					preparedSuperProcs = append(preparedSuperProcs, PreparedSuperProc{
						ProcInfo:         proc,
						Parent:           last.ProcInfo,
						ParentStructName: last.StructName,
					})
				}
				lastSeenProc[proc.Name] = PreparedProc{
					ProcInfo:   proc,
					StructName: source.StructName,
				}
			}
			encodeMap[source].ProcSupers = preparedSuperProcs
		}
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
		for _, source := range chunk.Sources {
			imports[source.Package] = struct{}{}
			for _, vi := range source.Vars {
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
	}
	delete(imports, tree.ImplImport)
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
		for _, source := range chunk.Sources {
			for _, vi := range source.Vars {
				vars = append(vars, PreparedVar{
					VarInfo:      vi,
					StructName:   source.StructName,
					PackageShort: source.PackageShort,
				})
			}
			for _, gi := range source.Getters {
				if hasgetter[gi.FieldName] {
					continue
				}
				hasgetter[gi.FieldName] = true
				if !gi.HasGetter {
					return nil, nil, nil, fmt.Errorf("no getter for field: %s", gi.FieldName)
				}
				getters = append(getters, PreparedGetter{
					GetterInfo: *gi,
					StructName: source.StructName,
				})
			}
			for _, pi := range source.Procs {
				if hasproc[pi.Name] {
					continue
				}
				hasproc[pi.Name] = true
				procs = append(procs, PreparedProc{
					ProcInfo:   pi,
					StructName: source.StructName,
				})
			}
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

func (source *SourceInfo) Encode(tree *TreeInfo, t *TypeInfo) (*PreparedChunk, error) {
	if !source.FoundConstructor {
		return nil, fmt.Errorf("no constructor for %s source %s.%s", t.Path, source.Package, source.StructName)
	}
	return &PreparedChunk{
		PackageShort: source.PackageShort,
		Package:      source.Package,
		StructName:   source.StructName,
		Vars:         source.Vars,
		// ProcSupers is populated later
	}, nil
}

func (p *PreparedImplementation) RevChunks() []*PreparedChunk {
	nchunks := make([]*PreparedChunk, len(p.Chunks))
	for i := 0; i < len(nchunks); i++ {
		nchunks[i] = p.Chunks[len(nchunks)-1-i]
	}
	return nchunks
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
			TypePath:    typeInfo.Path,
			ImplPackage: t.ImplPackage,
			Type:        typeInfo.Type(),
			Imports:     imports,
			Chunks:      chunks,
			ParentPath:  parent,
			Vars:        vars,
			Procs:       procs,
			Getters:     getters,
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
