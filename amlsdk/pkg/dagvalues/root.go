package dagvalues

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"dagger.io/dagger/querybuilder"
	"github.com/dagger/dagger/cmd/codegen/generator"
)

func Dagger(ctx context.Context, dag *dagger.Client) (Object, error) {
	schema, err := generator.Introspect(ctx, dag)
	if err != nil {
		return Object{}, err
	}

	query := schema.Types.Get("Query")
	if query == nil {
		return Object{}, fmt.Errorf("failed to find query")
	}

	return Object{
		selection:   querybuilder.Query(),
		dag:         dag,
		schema:      schema,
		currentType: query,
	}, nil
}
