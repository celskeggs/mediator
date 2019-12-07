package main

import (
	"errors"
	"fmt"
	"github.com/celskeggs/mediator/util"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path"
	"strings"
	"unicode"
)

func ToSnakeCase(name string) string {
	var out []rune
	for _, r := range name {
		if unicode.IsUpper(r) {
			if len(out) > 0 {
				out = append(out, '_')
			}
			r = unicode.ToLower(r)
		}
		out = append(out, r)
	}
	return string(out)
}

func (info *TreeInfo) Dump() error {
	for _, t := range info.Paths {
		if t.FoundConstructor {
			fmt.Printf("FOUND CONSTRUCTOR FOR %s\n", t.Path)
		}
		for _, vi := range t.Vars {
			fmt.Printf("FIELD: %s for %s\n", vi.FieldName, t.Path)
			err := ast.Print(vi.FileSet, vi.Type)
			if err != nil {
				return err
			}
		}
		for _, gi := range t.Getters {
			if gi.HasGetter {
				fmt.Printf("GETTER: %s for %s\n", gi.FieldName, t.Path)
			}
			if gi.HasSetter {
				fmt.Printf("SETTER: %s for %s\n", gi.FieldName, t.Path)
			}
		}
		for _, p := range t.Procs {
			fmt.Printf("PROC: %s for %s with %d parameters\n", p.Name, t.Path, p.ParamCount)
		}
	}
	return nil
}

func (info *TypeInfo) LoadStruct(fset *token.FileSet, structType *ast.StructType, context *ast.File) error {
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			if strings.HasPrefix(name.Name, "Var") {
				info.Vars = append(info.Vars, VarInfo{
					FieldName: ToSnakeCase(name.Name[3:]),
					LongName:  name.Name,
					Type:      field.Type,
					FileSet:   fset,

					DefiningImports: context.Imports,
				})
			}
		}
	}
	return nil
}

func IsDatumType(ref ast.Expr) bool {
	star, ok := ref.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	util.FIXME("check that imports are also correct")
	if !ok || ident.Name != "types" {
		return false
	}
	if selector.Sel.Name != "Datum" {
		return false
	}
	return true
}

func IsValueType(ref ast.Expr) bool {
	selector, ok := ref.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	util.FIXME("check that imports are also correct")
	if !ok || ident.Name != "types" {
		return false
	}
	if selector.Sel.Name != "Value" {
		return false
	}
	return true
}

func (info *TypeInfo) LoadNewFunc(fset *token.FileSet, decl *ast.FuncDecl) error {
	// can assume that name is correct
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		return fmt.Errorf("constructor for %s must be global", info.StructName)
	}
	if len(decl.Type.Results.List) != 1 {
		return fmt.Errorf("constructor for %s must return exactly one value", info.StructName)
	}
	resultType := decl.Type.Results.List[0].Type
	ident, ok := resultType.(*ast.Ident)
	if !ok || ident.Name != info.StructName {
		return fmt.Errorf("constructor for %s must return plain structure result", info.StructName)
	}
	if len(decl.Type.Params.List) != 2 {
		return fmt.Errorf("constructor for %s must accept exactly two parameters", info.StructName)
	}
	if !IsDatumType(decl.Type.Params.List[0].Type) {
		return fmt.Errorf("constructor for %s must accept a datum parameter", info.StructName)
	}
	paramType2 := decl.Type.Params.List[1].Type
	elt, ok := paramType2.(*ast.Ellipsis)
	if !ok {
		return fmt.Errorf("constructor for %s must accept a varargs parameter", info.StructName)
	}
	if !IsValueType(elt.Elt) {
		return fmt.Errorf("constructor for %s must accept varargs of types.Value", info.StructName)
	}
	info.FoundConstructor = true
	return nil
}

func (info *TypeInfo) LoadProc(fset *token.FileSet, decl *ast.FuncDecl, name string) error {
	// can assume that receiver was already checked and that 'name' is the name of the proc
	var types []ast.Expr
	for _, param := range decl.Type.Params.List {
		types = append(types, param.Type)
	}
	if len(types) < 1 {
		return fmt.Errorf("proc %s.%s must take at least src at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
	}
	for i, t := range types {
		if i == 0 {
			if !IsDatumType(t) {
				return fmt.Errorf("proc %s.%s must take src from *types.Datum at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
			}
		} else {
			if !IsValueType(t) {
				return fmt.Errorf("proc %s.%s must take only types.Value at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
			}
		}
	}
	if decl.Type.Results != nil && len(decl.Type.Results.List) > 1 {
		return fmt.Errorf("proc %s.%s cannot have more than one result at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
	}
	var resultType ast.Expr
	if decl.Type.Results != nil && len(decl.Type.Results.List) > 0 {
		resultType = decl.Type.Results.List[0].Type
	}
	if !IsValueType(resultType) {
		return fmt.Errorf("proc %s.%s must return a types.Value at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
	}
	info.Procs = append(info.Procs, ProcInfo{
		Name:       name,
		ParamCount: len(types) - 1,
	})
	return nil
}

func (info *TypeInfo) GetterInfo(name string) *GetterInfo {
	for _, gi := range info.Getters {
		if gi.LongName == name {
			return gi
		}
	}
	gi := &GetterInfo{
		FieldName: ToSnakeCase(name),
		LongName:  name,
		HasGetter: false,
		HasSetter: false,
	}
	info.Getters = append(info.Getters, gi)
	return gi
}

func (info *TypeInfo) LoadGetter(fset *token.FileSet, decl *ast.FuncDecl) error {
	// can assume that receiver was already checked and that name starts with Get but is longer
	if decl.Type.Results == nil || len(decl.Type.Results.List) != 1 {
		return fmt.Errorf("getter %s.%s must return exactly one value at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
	}
	if !IsValueType(decl.Type.Results.List[0].Type) {
		return fmt.Errorf("getter %s.%s must return a types.Value at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
	}
	if decl.Type.Params == nil || len(decl.Type.Params.List) != 1 {
		return fmt.Errorf("getter %s.%s must accept exactly one parameter at %v", info.StructName, decl.Name.Name, fset.Position(decl.Type.Pos()))
	}
	if !IsDatumType(decl.Type.Params.List[0].Type) {
		return fmt.Errorf("getter %s.%s must accept a *types.Datum at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
	}
	util.FIXME("make sure that vars and getters don't conflict")
	gi := info.GetterInfo(decl.Name.Name[3:])
	if gi.HasGetter {
		return fmt.Errorf("duplicate getter %s.%s at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
	}
	gi.HasGetter = true
	return nil
}

func (info *TypeInfo) LoadSetter(fset *token.FileSet, decl *ast.FuncDecl) error {
	// can assume that receiver was already checked and that name starts with Set but is longer
	if decl.Type.Results != nil && len(decl.Type.Results.List) != 0 {
		return fmt.Errorf("setter %s.%s must return no value at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
	}
	if len(decl.Type.Params.List) != 2 {
		return fmt.Errorf("setter %s.%s must accept exactly two parameters at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
	}
	if !IsDatumType(decl.Type.Params.List[0].Type) {
		return fmt.Errorf("setter %s.%s must accept a *types.Datum at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
	}
	if !IsValueType(decl.Type.Params.List[1].Type) {
		return fmt.Errorf("setter %s.%s must accept a types.Value at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
	}
	gi := info.GetterInfo(decl.Name.Name[3:])
	if gi.HasSetter {
		return fmt.Errorf("duplicate setter %s.%s at %v", info.StructName, decl.Name.Name, fset.Position(decl.Pos()))
	}
	gi.HasSetter = true
	return nil
}

func (t *TreeInfo) LoadAST(fset *token.FileSet, file *ast.File, importPath string) error {
	for _, decl := range file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok {
			if gen.Tok == token.TYPE {
				if len(gen.Specs) != 1 {
					return errors.New("expected exactly one spec")
				}
				spec := gen.Specs[0].(*ast.TypeSpec)
				for _, ti := range t.Paths {
					if ti.StructName == spec.Name.Name && ti.Package == importPath {
						stype, ok := spec.Type.(*ast.StructType)
						if !ok {
							return errors.New("expected struct type")
						}
						err := ti.LoadStruct(fset, stype, file)
						if err != nil {
							return err
						}
					}
				}
			}
		} else if fun, ok := decl.(*ast.FuncDecl); ok {
			for _, ti := range t.Paths {
				if fun.Name.Name == "New"+ti.StructName && ti.Package == importPath {
					err := ti.LoadNewFunc(fset, fun)
					if err != nil {
						return err
					}
				}
				if fun.Recv != nil && len(fun.Recv.List) == 1 {
					recvType := fun.Recv.List[0].Type
					if star, ok := recvType.(*ast.StarExpr); ok {
						recvType = star.X
					}
					if ident, ok := recvType.(*ast.Ident); ok {
						if ident.Name == ti.StructName && ti.Package == importPath {
							if strings.HasPrefix(fun.Name.Name, "Proc") && len(fun.Name.Name) > len("Proc") {
								err := ti.LoadProc(fset, fun, fun.Name.Name[4:])
								if err != nil {
									return err
								}
							} else if strings.HasPrefix(fun.Name.Name, "Get") && len(fun.Name.Name) > len("Get") {
								err := ti.LoadGetter(fset, fun)
								if err != nil {
									return err
								}
							} else if strings.HasPrefix(fun.Name.Name, "Set") && len(fun.Name.Name) > len("Set") {
								err := ti.LoadSetter(fset, fun)
								if err != nil {
									return err
								}
							} else if fun.Name.Name == "OperatorWrite" {
								err := ti.LoadProc(fset, fun, "<<")
								if err != nil {
									return err
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func (t *TreeInfo) LoadPackages() error {
	fset := token.NewFileSet()
	for _, pkg := range t.Packages {
		for _, filename := range pkg.GoFiles {
			fileAST, err := parser.ParseFile(fset, path.Join(pkg.Dir, filename), nil, parser.AllErrors)
			if err != nil {
				return err
			}
			err = t.LoadAST(fset, fileAST, pkg.ImportPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *TreeInfo) NewType(path string) (*TypeInfo, error) {
	if _, exists := t.Paths[path]; exists {
		return nil, fmt.Errorf("double declaration of %s", path)
	}
	ti := &TypeInfo{Path: path}
	t.Paths[path] = ti
	return ti, nil
}

func (t *TreeInfo) LoadFromDecl(decl Decl) error {
	nt, err := t.NewType(decl.Path)
	if err != nil {
		return err
	}
	nt.Parent = decl.ParentPath
	nt.StructName = decl.StructName
	nt.Package = decl.Package.ImportPath
	nt.PackageShort, err = t.GetPackageName(nt.Package)
	if err != nil {
		return err
	}
	alreadyExists := false
	for _, pkg := range t.Packages {
		if pkg.Dir == decl.Package.Dir {
			alreadyExists = true
			break
		}
	}
	if !alreadyExists {
		t.Packages = append(t.Packages, decl.Package)
	}
	return nil
}

func (t *TreeInfo) GetPackageName(importPath string) (string, error) {
	if val, ok := t.PkgNames[importPath]; ok {
		return val, nil
	}
	pkg, err := build.Default.Import(importPath, "", 0)
	if err != nil {
		return "", err
	}
	t.PkgNames[importPath] = pkg.Name
	return pkg.Name, nil
}
