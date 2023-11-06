package dagvalues

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"dagger.io/dagger/querybuilder"
	"github.com/acorn-io/aml/pkg/value"
	"github.com/dagger/dagger/cmd/codegen/introspection"
)

var _ value.Schema = Type{}

type Type struct {
	dag         *dagger.Client
	schema      *introspection.Schema
	currentType *introspection.Type
}

func (t Type) Eq(right value.Value) (value.Value, error) {
	if rightType, ok := right.(Type); ok {
		return value.NewValue(t.currentType.Name == rightType.currentType.Name), nil
	}
	return value.False, nil
}

func (t Type) Kind() value.Kind {
	return "dagger:type"
}

func (t Type) MergeType(right value.Schema) (value.Schema, error) {
	return right, nil
}

func (t Type) Default() (value.Value, bool, error) {
	return nil, false, nil
}

func (t Type) DefaultWithImplicit(implicit bool) (value.Value, bool, error) {
	return nil, false, nil
}

func (t Type) Validate(ctx context.Context, v value.Value) (value.Value, error) {
	if t.TargetKind() == v.Kind() {
		return v, nil
	} else if v.Kind() == value.ObjectKind {
		return v, nil
	}

	currentType := t.schema.Types.Get(t.currentType.Name)
	if currentType == nil {
		return Object{}, fmt.Errorf("failed to find dagger type %s", t.currentType.Name)
	}

	return &Object{
		selection: querybuilder.Query().
			Select(fmt.Sprintf("load%sFromID", t.currentType.Name)).
			Arg("id", fmt.Sprint(v)),
		dag:         t.dag,
		schema:      t.schema,
		currentType: currentType,
	}, nil
}

func (t Type) TargetCompatible(target value.Value) bool {
	kind := target.Kind()
	return kind == value.ObjectKind ||
		kind == value.StringKind
}

func (t Type) TargetKind() value.Kind {
	return value.Kind(t.GetPath().String())
}

func (t Type) ValidArrayItems() []value.Schema {
	return nil
}

func (t Type) GetPath() value.Path {
	dag := "dag"
	return value.Path{
		value.PathElement{
			Key: &dag,
		},
		value.PathElement{
			Key: &t.currentType.Name,
		},
	}
}
