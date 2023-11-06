package dagvalues

import (
	"context"

	"github.com/acorn-io/aml/pkg/value"
)

type Scalar struct {
	f Function
}

func (s Scalar) NativeValue() (any, bool, error) {
	v, ok, err := value.Call(context.Background(), s)
	if err != nil || !ok {
		return nil, ok, err
	}
	return value.NativeValue(v)
}

func (s Scalar) Call(ctx context.Context, args []value.CallArgument) (value.Value, bool, error) {
	newQuery, err := s.f.buildObject(args)
	if err != nil {
		return nil, false, err
	}

	var (
		result   value.Value
		q        = newQuery.selection
		typeName = newQuery.currentType.Name
	)

	// There's got to be a smarter way to do this
	switch typeName {
	case "Boolean":
		var b bool
		q = q.Bind(&b)
		err = q.Execute(ctx, newQuery.dag.GraphQLClient())
		result = value.NewValue(b)
	case "Int":
		var i int64
		q = q.Bind(&i)
		err = q.Execute(ctx, newQuery.dag.GraphQLClient())
		result = value.NewValue(i)
	case "Float":
		var i float64
		q = q.Bind(&i)
		err = q.Execute(ctx, newQuery.dag.GraphQLClient())
		result = value.NewValue(i)
	case "List":
		var l []any
		q = q.Bind(&l)
		err = q.Execute(ctx, newQuery.dag.GraphQLClient())
		result = value.NewValue(l)
	default:
		var s string
		q = q.Bind(&s)
		err = q.Execute(ctx, newQuery.dag.GraphQLClient())
		result = value.NewValue(s)
	}
	return result, true, err
}

func (s Scalar) Kind() value.Kind {
	return "dagger:scalar"
}
