package lib

import (
	"strings"
	"fmt"
	"github.com/celskeggs/mediator/autocoder/iface"
	"github.com/celskeggs/mediator/autocoder/gotype"
)

type varNameCtx struct {
	Next map[rune]int
}

func newVarNameCtx() *varNameCtx {
	return &varNameCtx{
		Next: map[rune]int{},
	}
}

func (v *varNameCtx) Name(base string) string {
	if base == "" {
		base = "var"
	} else {
		base = strings.ToLower(base)
	}
	r := []rune(base)[0]
	cur := v.Next[r]
	v.Next[r] += 1
	if cur == 0 {
		return base[0:1]
	} else {
		return fmt.Sprintf("%s%d", base[0:1], cur)
	}
}

func (v *varNameCtx) Var(base string) iface.Expr {
	return formatExpression("%s", v.Name(base))
}

func (v *varNameCtx) VarFromType(base gotype.Type) iface.Expr {
	return v.Var(base.Name())
}

func (v *varNameCtx) VarsFromTypes(basis []gotype.Type) (paramStr string, exprs []iface.Expr) {
	var paramStrs []string
	for _, base := range basis {
		expr := v.VarFromType(base)
		paramStrs = append(paramStrs, fmt.Sprintf("%v %v", expr, base))
		exprs = append(exprs, expr)
	}
	return strings.Join(paramStrs, ", "), exprs
}
