package convert

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/gen"
	"github.com/celskeggs/mediator/autocoder/pack"
	"github.com/celskeggs/mediator/dream/ast"
	"github.com/celskeggs/mediator/dream/parser"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/pkg/errors"
)

func Convert(dmf *ast.File, packageName string) (*gen.DefinedTree, error) {
	dt := &gen.DefinedTree{
		Package:   packageName,
		WorldMob:  path.ConstTypePath("/mob"),
		WorldName: "World",
		Maps:      dmf.Maps,
	}
	// define all types
	for _, def := range dmf.Definitions {
		if def.Type == ast.DefTypeDefine {
			err := DefinePath(dt, def.Path)
			if err != nil {
				return nil, err
			}
		}
	}
	// declare all variables, procedures, and verbs
	for _, def := range dmf.Definitions {
		var err error
		if def.Type == ast.DefTypeVarDef {
			err = DefineVar(dt, def.Path, def.VarType, def.Variable, def.SourceLoc)
		} else if def.Type == ast.DefTypeProcDecl {
			err = DefineProc(dt, def.Path, false, def.Variable, def.SourceLoc)
		} else if def.Type == ast.DefTypeVerbDecl {
			err = DefineProc(dt, def.Path, true, def.Variable, def.SourceLoc)
		}
		if err != nil {
			return nil, err
		}
	}
	// assign all values
	for _, def := range dmf.Definitions {
		if def.Type == ast.DefTypeAssign {
			err := AssignPath(dt, def.Path, def.Variable, def.Expression, def.SourceLoc)
			if err != nil {
				return nil, err
			}
		}
	}
	// implement all functions
	for _, def := range dmf.Definitions {
		if def.Type == ast.DefTypeImplement {
			err := ImplementFunction(dt, def.Path, def.Variable, def.Arguments, def.Body, def.SourceLoc)
			if err != nil {
				return nil, err
			}
		}
	}
	// insert names for everything unnamed
	for i, t := range dt.Types {
		if dt.Extends(t.TypePath, path.ConstTypePath("/atom")) && !t.IsOverride() {
			specifiesName := false
			for _, init := range t.Inits {
				if init.Name == "name" {
					specifiesName = true
				}
			}
			if !specifiesName {
				_, lastComponent, err := t.TypePath.SplitLast()
				if err != nil {
					return nil, err
				}
				t.Inits = append(t.Inits, gen.DefinedInit{
					Name:      "name",
					Value:     fmt.Sprintf("types.String(%q)", lastComponent),
					SourceLoc: tokenizer.SourceHere(),
				})
				dt.Types[i] = t
			}
		}
	}
	return dt, nil
}

func ConvertFiles(inputFiles []string, outputGo string, outputPack string, packageName string) error {
	dmf, err := parser.ParseFiles(inputFiles)
	if err != nil {
		return errors.Wrap(err, "while parsing input files")
	}
	tree, err := Convert(dmf, packageName)
	if err != nil {
		return errors.Wrap(err, "while building tree")
	}
	err = gen.GenerateTo(tree, outputGo)
	if err != nil {
		return errors.Wrap(err, "while generating output file")
	}
	err = pack.GenerateResourcePack(dmf, outputPack)
	if err != nil {
		return errors.Wrap(err, "while generating resource pack")
	}
	return nil
}
