package mysql

import (
	"context"
	"io"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/mysql/dialect"
	"github.com/stephenafamo/bob/expr"
	"github.com/stephenafamo/bob/mods"
)

type Expression = dialect.Expression

//nolint:gochecknoglobals
var bmod = expr.Builder[Expression, Expression]{}

// F creates a function expression with the given name and args
//
//	SQL: generate_series(1, 3)
//	Go: mysql.F("generate_series", 1, 3)
func F(name string, args ...any) mods.Moddable[*dialect.Function] {
	f := dialect.NewFunction(name, args...)

	return mods.Moddable[*dialect.Function](func(mods ...bob.Mod[*dialect.Function]) *dialect.Function {
		for _, mod := range mods {
			mod.Apply(f)
		}

		return f
	})
}

// S creates a string literal
// SQL: 'a string'
// Go: mysql.S("a string")
func S(s string) Expression {
	return bmod.S(s)
}

// SQL: NOT true
// Go: mysql.Not("true")
func Not(exp bob.Expression) Expression {
	return bmod.Not(exp)
}

// SQL: a OR b OR c
// Go: mysql.Or("a", "b", "c")
func Or(args ...bob.Expression) Expression {
	return bmod.Or(args...)
}

// SQL: a AND b AND c
// Go: mysql.And("a", "b", "c")
func And(args ...bob.Expression) Expression {
	return bmod.And(args...)
}

// SQL: a || b || c
// Go: mysql.Concat("a", "b", "c")
func Concat(args ...bob.Expression) Expression {
	return expr.X[Expression, Expression](expr.Join{Exprs: args, Sep: " || "})
}

// SQL: $1, $2, $3
// Go: mysql.Args("a", "b", "c")
func Arg(args ...any) Expression {
	return bmod.Arg(args...)
}

// SQL: ($1, $2, $3)
// Go: mysql.ArgGroup("a", "b", "c")
func ArgGroup(args ...any) Expression {
	return bmod.ArgGroup(args...)
}

// SQL: $1, $2, $3
// Go: mysql.Placeholder(3)
func Placeholder(n uint) Expression {
	return bmod.Placeholder(n)
}

// SQL: (a, b)
// Go: mysql.Group("a", "b")
func Group(exps ...bob.Expression) Expression {
	return bmod.Group(exps...)
}

// SQL: "table"."column"
// Go: mysql.Quote("table", "column")
func Quote(ss ...string) Expression {
	return bmod.Quote(ss...)
}

// SQL: where a = $1
// Go: mysql.Raw("where a = ?", "something")
func Raw(query string, args ...any) Expression {
	return bmod.Raw(query, args...)
}

// SQL: CAST(a AS int)
// Go: psql.Cast("a", "int")
func Cast(exp bob.Expression, typname string) Expression {
	return bmod.Cast(exp, typname)
}

type CaseChain[T bob.Expression] func() expr.Case[T]

func (c CaseChain[T]) WriteSQL(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
	return c().WriteSQL(ctx, w, d, start)
}

func Case() CaseChain[Expression] {
	return CaseChain[Expression](func() expr.Case[Expression] { return expr.Case[Expression]{} })
}

func (c CaseChain[T]) When(condition, then T) CaseChain[T] {
	cExpr := c()
	cExpr.Whens = append(cExpr.Whens, expr.When{Condition: condition, Then: then})
	return CaseChain[T](func() expr.Case[T] { return cExpr })
}

func (c CaseChain[T]) Else(then T) Expression {
	cExpr := c()
	cExpr.Else = then
	var e dialect.Expression
	return e.New(cExpr)
}

// func (c CaseChain[T]) As(alias string) T {
// 	return bmod.X(c())
// }
