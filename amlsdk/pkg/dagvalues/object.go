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

func (o Object) Eq(right value.Value) (value.Value, error) {
	return value.NewValue(o.Kind() == right.Kind()), nil
}

func (o Object) String() string {
	return string(o.Kind())
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
		if field.Name != name {
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

	returnType := o.schema.Types.Get(name)
	if returnType != nil {
		return Type{
			dag:         o.dag,
			schema:      o.schema,
			currentType: returnType,
		}, true, nil
	}

	return nil, false, nil
}

func (o Object) Compatible(kind value.Kind) bool {
	return kind == value.ObjectKind ||
		kind == value.StringKind ||
		strings.HasPrefix(string(kind), "dag.")
}

func (o Object) Kind() value.Kind {
	return value.Kind("dag." + o.currentType.Name)
}
