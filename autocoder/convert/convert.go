package convert

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/autocoder/gen"
	"github.com/celskeggs/mediator/dream/ast"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/celskeggs/mediator/util"
)

func DefinePath(dt *gen.DefinedTree, path path.TypePath) error {
	if dt.GetTypeByPath(path) != nil {
		return nil
	}
	switch path.String() {
	case "/world":
		// nothing to do
	default:
		if dt.Exists(path) {
			// exists, but not as a locally-defined type: we're trying to override something!
			dt.Types = append(dt.Types, gen.DefinedType{
				TypePath: path,
				BasePath: path,
			})
		} else {
			dt.Types = append(dt.Types, gen.DefinedType{
				TypePath: path,
			})
		}
	}
	return nil
}

func DefineVar(dt *gen.DefinedTree, path path.TypePath, varType path.TypePath, variable string, loc tokenizer.SourceLocation) error {
	if !dt.Exists(path) {
		return fmt.Errorf("no such path %v for declaration of variable %v at %v", path, variable, loc)
	}
	defType := dt.GetTypeByPath(path)
	if defType == nil {
		panic("expected non-nil type " + path.String())
	}
	if !varType.IsEmpty() {
		return fmt.Errorf("unimplemented: variable type %v at %v", varType, loc)
	}

	_, found := dt.ResolveField(path, variable)
	if found {
		return fmt.Errorf("field %s already defined on %v at %v", variable, path, loc)
	}

	defType.Fields = append(defType.Fields, gen.DefinedField{
		Name: variable,
		Type: dtype.Any(),
	})
	return nil
}

func DefineProc(dt *gen.DefinedTree, typePath path.TypePath, isVerb bool, variable string, loc tokenizer.SourceLocation) error {
	if !dt.Exists(typePath) {
		return fmt.Errorf("no such path %v for declaration of proc/verb %v at %v", typePath, variable, loc)
	}
	defType := dt.GetTypeByPath(typePath)
	if defType == nil {
		panic("expected non-nil type " + typePath.String())
	}
	if isVerb {
		if !dt.Extends(typePath, path.ConstTypePath("/atom")) {
			return fmt.Errorf("attempt to declare verb %v on non-atom %v at %v", variable, typePath, loc)
		}
		defType.Verbs = append(defType.Verbs, variable)
	}

	_, found := dt.ResolveProcedure(typePath, variable)
	if found {
		return fmt.Errorf("proc/verb %s already defined on %v at %v", variable, typePath, loc)
	}

	defType.Procs = append(defType.Procs, gen.DefinedProc{
		Name: variable,
	})
	return nil
}

func AssignPath(dt *gen.DefinedTree, path path.TypePath, variable string, expr ast.Expression, loc tokenizer.SourceLocation) error {
	switch path.String() {
	case "/world":
		switch variable {
		case "name":
			dt.WorldName = ConstantString(expr)
		case "mob":
			dt.WorldMob = ConstantPath(expr)
			if !dt.Exists(dt.WorldMob) {
				panic("path " + dt.WorldMob.String() + " does not actually exist in the tree")
			}
		default:
			return fmt.Errorf("no such path %v for assignment of variable %v", path, variable)
		}
	default:
		if !dt.Exists(path) {
			return fmt.Errorf("no such path %v for assignment of variable %v", path, variable)
		}
		util.FIXME("some sort of typechecking for field assignments?")
		_, found := dt.ResolveField(path, variable)
		if !found {
			return fmt.Errorf("no such field %s on %v at %v", variable, path, loc)
		}
		util.FIXME("make sure that users can't initialize 'verbs' field without making that work out")
		defType := dt.GetTypeByPath(path)
		util.FIXME("make it easier to tell that atoms.WorldOf(src) is valid here")
		expr, _, err := ExprToGo(expr, CodeGenContext{
			WorldRef: "atoms.WorldOf(src)",
			Tree:     dt,
		})
		if err != nil {
			return err
		}
		defType.Inits = append(defType.Inits, gen.DefinedInit{
			Name:      variable,
			Value:     expr,
			SourceLoc: loc,
		})
	}
	return nil
}

func ImplementFunction(dt *gen.DefinedTree, path path.TypePath, function string, arguments []ast.TypedName, body []ast.Statement, loc tokenizer.SourceLocation) error {
	if !dt.Exists(path) {
		return fmt.Errorf("no such path %v for implementation of function %v at %v", path, function, loc)
	}
	defType := dt.GetTypeByPath(path)
	if defType == nil {
		panic("expected non-nil type " + path.String())
	}

	_, found := dt.ResolveProcedure(path, function)
	if !found {
		return fmt.Errorf("no such function %s to implement on %v at %v", function, path, loc)
	}

	var params []string
	vartypes := map[string]dtype.DType{}
	vartypes["src"] = dtype.Path(path)
	vartypes["usr"] = dtype.ConstPath("/mob")
	for _, a := range arguments {
		params = append(params, LocalVariablePrefix+a.Name)
		vartypes[a.Name] = a.Type
	}

	settings, body, err := ParseSettings(dt, path, body)
	if err != nil {
		return err
	}

	lines, err := FuncBodyToGo(body, CodeGenContext{
		WorldRef: "atoms.WorldOf(" + LocalVariablePrefix + "src)",
		Tree:     dt,
		VarTypes: vartypes,
		Result:   "out",
	})
	if err != nil {
		return err
	}

	defType.Impls = append(defType.Impls, gen.DefinedImpl{
		Name:     function,
		This:     LocalVariablePrefix + "src",
		Usr:      LocalVariablePrefix + "usr",
		Params:   params,
		Settings: settings,
		Body:     MergeGoLines(lines),
	})
	return nil
}
