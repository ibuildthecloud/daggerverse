package dagvalues

import (
	"context"
	"fmt"

	"github.com/acorn-io/aml/pkg/value"
	"github.com/dagger/dagger/cmd/codegen/introspection"
)

type Function struct {
	obj   Object
	field *introspection.Field
}

func (f Function) LookupValue(key value.Value) (value.Value, bool, error) {
	v, ok, err := f.Call(context.Background(), nil)
	if err != nil || !ok {
		return nil, ok, err
	}
	return value.Lookup(v, key)
}

func (f Function) String() string {
	return "dagger:type:function:" + f.obj.currentType.Name + ":" + f.field.Name
}

func (f Function) Merge(right value.Value) (value.Value, error) {
	v, _, err := f.Call(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return value.Merge(v, right)
}

func (f Function) Call(_ context.Context, args []value.CallArgument) (value.Value, bool, error) {
	v, err := f.buildObject(args)
	return v, err == nil, err
}

func (f Function) buildObject(args []value.CallArgument) (Object, error) {
	argsValue, err := fieldArgsToValue(f.field, args)
	if err != nil {
		return Object{}, err
	}

	entries, err := value.Entries(argsValue)
	if err != nil {
		return Object{}, err
	}

	result := Object{
		selection:   f.obj.selection.Select(f.field.Name),
		dag:         f.obj.dag,
		schema:      f.obj.schema,
		currentType: f.obj.schema.Types.Get(realType(f.field.TypeRef).Name),
	}

	if result.currentType == nil && realType(f.field.TypeRef).Kind == introspection.TypeKindList {
		result.currentType = &introspection.Type{
			Name: "List",
		}
	}

	if result.currentType == nil {
		return Object{}, fmt.Errorf("failed to find return type")
	}

	for _, entry := range entries {
		var argValue any
		if t, ok := entry.Value.(Object); ok {
			argValue = t
		} else {
			nv, ok, err := value.NativeValue(entry.Value)
			if err != nil {
				return Object{}, err
			} else if !ok {
				continue
			}
			argValue = nv
		}

		for _, arg := range f.field.Args {
			if arg.Name == entry.Key && realType(arg.TypeRef).Kind == introspection.TypeKindEnum {
				argValue = Enum(argValue.(string))
			}
		}

		result.selection = result.selection.Arg(entry.Key, argValue)
	}

	return result, nil
}

func (f Function) Kind() value.Kind {
	return "dagger:function"
}

type Enum string

func (e Enum) IsEnum() {}
