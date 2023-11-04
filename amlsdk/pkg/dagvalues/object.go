package dagvalues

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"dagger.io/dagger/querybuilder"
	"github.com/acorn-io/aml/pkg/value"
	"github.com/dagger/dagger/cmd/codegen/introspection"
)

var _ querybuilder.GraphQLMarshaller = Object{}

type Object struct {
	selection   *querybuilder.Selection
	dag         *dagger.Client
	schema      *introspection.Schema
	currentType *introspection.Type
}

func (o Object) Keys() (result []string, _ error) {
	for _, field := range o.currentType.Fields {
		if isScalar(field.TypeRef) {
			result = append(result, field.Name)
		}
	}
	return
}

func (o Object) XXX_GraphQLType() string {
	return o.currentType.Name
}

func (o Object) MarshalJSON() ([]byte, error) {
	id, err := o.XXX_GraphQLID(context.Background())
	if err != nil {
		return nil, err
	}
	return json.Marshal(id)
}

func (o Object) XXX_GraphQLIDType() string {
	for _, field := range o.currentType.Fields {
		if field.Name == "id" {
			return string(realType(field.TypeRef).Kind)
		}
	}

	return "String"
}

func (o Object) XXX_GraphQLID(ctx context.Context) (string, error) {
	id, ok, err := value.Lookup(o, value.NewValue("id"))
	if err != nil {
		return "", err
	} else if !ok {
		return "", fmt.Errorf("failed to find id scalar")
	}
	v, ok, err := value.Call(ctx, id)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("id() call returned no value")
	}
	return value.ToString(v)
}

func (o Object) String() string {
	return "dagger:type:" + o.currentType.Name
}

func (o Object) NativeValue() (any, bool, error) {
	id, err := o.XXX_GraphQLID(context.Background())
	if err != nil {
		return o.String() + ":" + err.Error(), true, nil
	}
	return id, true, nil
}

func (o Object) LookupValue(key value.Value) (value.Value, bool, error) {
	name, err := value.ToString(key)
	if err != nil {
		return nil, false, err
	}

	for _, field := range o.currentType.Fields {
		if !strings.EqualFold(field.Name, name) {
			continue
		}

		result := Function{
			obj:   o,
			field: field,
		}
		if isScalar(field.TypeRef) {
			return Scalar{
				f: result,
			}, true, nil
		}
		return result, true, nil
	}

	return nil, false, nil
}

func (o Object) Merge(right value.Value) (value.Value, error) {
	entries, err := value.Entries(right)
	if err != nil {
		return nil, err
	}

	result := value.Value(o)
	for _, entry := range entries {
		call, ok, err := value.Lookup(result, value.NewValue(entry.Key))
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if _, ok := call.(Scalar); ok {
			return nil, fmt.Errorf("scalars can not be defined as an object, only used for lookup, invalid key [%s]", entry.Key)
		}

		args := []value.CallArgument{{Value: entry.Value}}
		if entry.Value.Kind() != value.ObjectKind {
			args[0].Positional = true
		}

		newValue, ok, err := value.Call(context.Background(), call, args...)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		result = newValue
	}

	return result, nil
}

func (o Object) Kind() value.Kind {
	return value.ObjectKind
}
