package dagvalues

import (
	"context"

	"github.com/acorn-io/aml/pkg/value"
)

type Type struct {
	typeName string
}

func (t Type) Kind() value.Kind {
	return value.SchemaKind
}

func (t Type) Validate(ctx context.Context, v value.Value) (value.Value, error) {
	return v, nil
}

func (t Type) TargetKind() value.Kind {
	return value.ObjectKind
}

func (t Type) ValidArrayItems() []value.Schema {
	return nil
}

func (t Type) GetPath() string {
	return "dagger." + t.typeName
}
