package dialect

import (
	"io"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/clause"
	"github.com/stephenafamo/bob/expr"
	"github.com/stephenafamo/bob/mods"
)

//nolint:gochecknoglobals
var bmod = expr.Builder[Expression, Expression]{}

func With[Q interface{ AppendWith(clause.CTE) }](name string, columns ...string) CTEChain[Q] {
	return CTEChain[Q](func() clause.CTE {
		return clause.CTE{
			Name:    name,
			Columns: columns,
		}
	})
}

type fromable interface {
	SetTable(any)
	SetTableAlias(alias string, columns ...string)
	SetLateral(bool)
	AppendPartition(...string)
	AppendIndexHint(clause.IndexHint)
}

func From[Q fromable](table any) FromChain[Q] {
	return FromChain[Q](func() clause.From {
		return clause.From{
			Table: table,
		}
	})
}

type FromChain[Q fromable] func() clause.From

func (f FromChain[Q]) Apply(q Q) {
	from := f()

	q.SetTable(from.Table)
	if from.Alias != "" {
		q.SetTableAlias(from.Alias, from.Columns...)
	}

	q.SetLateral(from.Lateral)
	q.AppendPartition(from.Partitions...)
}

func (f FromChain[Q]) As(alias string, columns ...string) FromChain[Q] {
	fr := f()
	fr.Alias = alias
	fr.Columns = columns

	return FromChain[Q](func() clause.From {
		return fr
	})
}

func (f FromChain[Q]) Lateral() FromChain[Q] {
	fr := f()
	fr.Lateral = true

	return FromChain[Q](func() clause.From {
		return fr
	})
}

func (f FromChain[Q]) Partition(partitions ...string) FromChain[Q] {
	fr := f()
	fr.Partitions = append(fr.Partitions, partitions...)

	return FromChain[Q](func() clause.From {
		return fr
	})
}

func (f FromChain[Q]) index(Type, For, first string, others ...string) FromChain[Q] {
	fr := f()
	fr.IndexHints = append(fr.IndexHints, clause.IndexHint{
		Type:    Type,
		Indexes: append([]string{first}, others...),
		For:     For,
	})

	return FromChain[Q](func() clause.From {
		return fr
	})
}

func (f FromChain[Q]) UseIndex(first string, others ...string) FromChain[Q] {
	return f.index("USE", "", first, others...)
}

func (f FromChain[Q]) UseIndexForJoin(first string, others ...string) FromChain[Q] {
	return f.index("USE", "JOIN", first, others...)
}

func (f FromChain[Q]) UseIndexForOrderBy(first string, others ...string) FromChain[Q] {
	return f.index("USE", "ORDER BY", first, others...)
}

func (f FromChain[Q]) UseIndexForGroupBy(first string, others ...string) FromChain[Q] {
	return f.index("USE", "GROUP BY", first, others...)
}

func (f FromChain[Q]) IgnoreIndex(first string, others ...string) FromChain[Q] {
	return f.index("IGNORE", "", first, others...)
}

func (f FromChain[Q]) IgnoreIndexForJoin(first string, others ...string) FromChain[Q] {
	return f.index("IGNORE", "JOIN", first, others...)
}

func (f FromChain[Q]) IgnoreIndexForOrderBy(first string, others ...string) FromChain[Q] {
	return f.index("IGNORE", "ORDER BY", first, others...)
}

func (f FromChain[Q]) IgnoreIndexForGroupBy(first string, others ...string) FromChain[Q] {
	return f.index("IGNORE", "GROUP BY", first, others...)
}

func (f FromChain[Q]) ForceIndex(first string, others ...string) FromChain[Q] {
	return f.index("FORCE", "", first, others...)
}

func (f FromChain[Q]) ForceIndexForJoin(first string, others ...string) FromChain[Q] {
	return f.index("FORCE", "JOIN", first, others...)
}

func (f FromChain[Q]) ForceIndexForOrderBy(first string, others ...string) FromChain[Q] {
	return f.index("FORCE", "ORDER BY", first, others...)
}

func (f FromChain[Q]) ForceIndexForGroupBy(first string, others ...string) FromChain[Q] {
	return f.index("FORCE", "GROUP BY", first, others...)
}

func Partition[Q interface{ AppendPartition(...string) }](partitions ...string) bob.Mod[Q] {
	return mods.QueryModFunc[Q](func(q Q) {
		q.AppendPartition(partitions...)
	})
}

type IndexHintChain[Q interface{ AppendIndexHint(clause.IndexHint) }] struct {
	hint clause.IndexHint
}

func (i *IndexHintChain[Q]) Apply(q Q) {
	q.AppendIndexHint(i.hint)
}

func (i *IndexHintChain[Q]) ForJoin() *IndexHintChain[Q] {
	i.hint.For = "JOIN"
	return i
}

func (i *IndexHintChain[Q]) ForOrderBy() *IndexHintChain[Q] {
	i.hint.For = "ORDER BY"
	return i
}

func (i *IndexHintChain[Q]) ForGroupBy() *IndexHintChain[Q] {
	i.hint.For = "GROUP BY"
	return i
}

type JoinChain[Q interface{ AppendJoin(clause.Join) }] func() clause.Join

func (j JoinChain[Q]) Apply(q Q) {
	q.AppendJoin(j())
}

func (j JoinChain[Q]) As(alias string) JoinChain[Q] {
	jo := j()
	jo.Alias = alias

	return JoinChain[Q](func() clause.Join {
		return jo
	})
}

func (j JoinChain[Q]) Natural() bob.Mod[Q] {
	jo := j()
	jo.Natural = true

	return mods.Join[Q](jo)
}

func (j JoinChain[Q]) On(on ...any) bob.Mod[Q] {
	jo := j()
	jo.On = append(jo.On, on)

	return mods.Join[Q](jo)
}

func (j JoinChain[Q]) OnEQ(a, b any) bob.Mod[Q] {
	jo := j()
	jo.On = append(jo.On, bmod.X(a).EQ(b))

	return mods.Join[Q](jo)
}

func (j JoinChain[Q]) Using(using ...any) bob.Mod[Q] {
	jo := j()
	jo.Using = using

	return mods.Join[Q](jo)
}

type joinable interface{ AppendJoin(clause.Join) }

func InnerJoin[Q joinable](e any) JoinChain[Q] {
	return JoinChain[Q](func() clause.Join {
		return clause.Join{
			Type: clause.InnerJoin,
			To:   e,
		}
	})
}

func LeftJoin[Q joinable](e any) JoinChain[Q] {
	return JoinChain[Q](func() clause.Join {
		return clause.Join{
			Type: clause.LeftJoin,
			To:   e,
		}
	})
}

func RightJoin[Q joinable](e any) JoinChain[Q] {
	return JoinChain[Q](func() clause.Join {
		return clause.Join{
			Type: clause.RightJoin,
			To:   e,
		}
	})
}

func CrossJoin[Q joinable](e any) bob.Mod[Q] {
	return mods.Join[Q]{
		Type: clause.CrossJoin,
		To:   e,
	}
}

func StraightJoin[Q joinable](e any) bob.Mod[Q] {
	return mods.Join[Q]{
		Type: clause.StraightJoin,
		To:   e,
	}
}

type OrderBy[Q interface{ AppendOrder(clause.OrderDef) }] func() clause.OrderDef

func (s OrderBy[Q]) Apply(q Q) {
	q.AppendOrder(s())
}

func (o OrderBy[Q]) Asc() OrderBy[Q] {
	order := o()
	order.Direction = "ASC"

	return OrderBy[Q](func() clause.OrderDef {
		return order
	})
}

func (o OrderBy[Q]) Desc() OrderBy[Q] {
	order := o()
	order.Direction = "DESC"

	return OrderBy[Q](func() clause.OrderDef {
		return order
	})
}

func (o OrderBy[Q]) Collate(collation string) OrderBy[Q] {
	order := o()
	order.CollationName = collation

	return OrderBy[Q](func() clause.OrderDef {
		return order
	})
}

type CTEChain[Q interface{ AppendWith(clause.CTE) }] func() clause.CTE

func (c CTEChain[Q]) Apply(q Q) {
	q.AppendWith(c())
}

func (c CTEChain[Q]) As(q bob.Query) CTEChain[Q] {
	cte := c()
	cte.Query = q
	return CTEChain[Q](func() clause.CTE {
		return cte
	})
}

type LockChain[Q interface{ SetFor(clause.For) }] func() clause.For

func (l LockChain[Q]) Apply(q Q) {
	q.SetFor(l())
}

func (l LockChain[Q]) NoWait() LockChain[Q] {
	lock := l()
	lock.Wait = clause.LockWaitNoWait
	return LockChain[Q](func() clause.For {
		return lock
	})
}

func (l LockChain[Q]) SkipLocked() LockChain[Q] {
	lock := l()
	lock.Wait = clause.LockWaitSkipLocked
	return LockChain[Q](func() clause.For {
		return lock
	})
}

type WindowMod[Q interface{ AppendWindow(clause.NamedWindow) }] struct {
	Name string
	*WindowChain[*WindowMod[Q]]
}

func (w WindowMod[Q]) Apply(q Q) {
	q.AppendWindow(clause.NamedWindow{
		Name:       w.Name,
		Definition: w.def,
	})
}

type WindowChain[T any] struct {
	def  clause.WindowDef
	Wrap T
}

func (w *WindowChain[T]) From(name string) T {
	w.def.SetFrom(name)
	return w.Wrap
}

func (w *WindowChain[T]) PartitionBy(condition ...any) T {
	w.def.AddPartitionBy(condition...)
	return w.Wrap
}

func (w *WindowChain[T]) OrderBy(order ...any) T {
	w.def.AddOrderBy(order...)
	return w.Wrap
}

func (w *WindowChain[T]) Range() T {
	w.def.SetMode("RANGE")
	return w.Wrap
}

func (w *WindowChain[T]) Rows() T {
	w.def.SetMode("ROWS")
	return w.Wrap
}

func (w *WindowChain[T]) FromUnboundedPreceding() T {
	w.def.SetStart("UNBOUNDED PRECEDING")
	return w.Wrap
}

func (w *WindowChain[T]) FromPreceding(exp any) T {
	w.def.SetStart(bob.ExpressionFunc(
		func(w io.Writer, d bob.Dialect, start int) ([]any, error) {
			return bob.ExpressIf(w, d, start, exp, true, "", " PRECEDING")
		}),
	)
	return w.Wrap
}

func (w *WindowChain[T]) FromCurrentRow() T {
	w.def.SetStart("CURRENT ROW")
	return w.Wrap
}

func (w *WindowChain[T]) FromFollowing(exp any) T {
	w.def.SetStart(bob.ExpressionFunc(
		func(w io.Writer, d bob.Dialect, start int) ([]any, error) {
			return bob.ExpressIf(w, d, start, exp, true, "", " FOLLOWING")
		}),
	)
	return w.Wrap
}

func (w *WindowChain[T]) ToPreceding(exp any) T {
	w.def.SetEnd(bob.ExpressionFunc(
		func(w io.Writer, d bob.Dialect, start int) ([]any, error) {
			return bob.ExpressIf(w, d, start, exp, true, "", " PRECEDING")
		}),
	)
	return w.Wrap
}

func (w *WindowChain[T]) ToCurrentRow(count int) T {
	w.def.SetEnd("CURRENT ROW")
	return w.Wrap
}

func (w *WindowChain[T]) ToFollowing(exp any) T {
	w.def.SetEnd(bob.ExpressionFunc(
		func(w io.Writer, d bob.Dialect, start int) ([]any, error) {
			return bob.ExpressIf(w, d, start, exp, true, "", " FOLLOWING")
		}),
	)
	return w.Wrap
}

func (w *WindowChain[T]) ToUnboundedFollowing() T {
	w.def.SetEnd("UNBOUNDED FOLLOWING")
	return w.Wrap
}