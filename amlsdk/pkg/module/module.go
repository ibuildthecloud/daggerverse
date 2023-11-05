package module

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"dagger.io/dagger"
	"github.com/acorn-io/aml/pkg/eval"
	"github.com/acorn-io/aml/pkg/value"
)

func isPublic(name string) bool {
	if len(name) == 0 {
		return false
	}
	return strings.ToUpper(name[0:1]) == name[0:1]
}

func toTypeDef(ctx context.Context, dag *dagger.Client, typeDef value.Schema) (*dagger.TypeDef, error) {
	if typeDef == nil {
		return dag.TypeDef().WithKind(dagger.Voidkind).WithOptional(true), nil
	}

	kind := typeDef.TargetKind()

	if value.IsSimpleKind(kind) {
		switch kind {
		case value.StringKind:
			return dag.TypeDef().WithKind(dagger.Stringkind), nil
		case value.NumberKind:
			return dag.TypeDef().WithKind(dagger.Integerkind), nil
		case value.BoolKind:
			return dag.TypeDef().WithKind(dagger.Booleankind), nil
		}
	}

	if kind == value.ArrayKind {
		valid := typeDef.ValidArrayItems()
		if len(valid) == 0 {
			return dag.TypeDef().WithListOf(dag.TypeDef().WithKind(dagger.Objectkind)), nil
		}

		itemType, err := toTypeDef(ctx, dag, valid[0])
		if err != nil {
			return nil, err
		}

		return dag.TypeDef().WithListOf(itemType), nil
	}

	typeName := typeDef.GetPath()
	if strings.HasPrefix(typeName, "dagger.") {
		typeName = strings.TrimPrefix(typeName, "dagger.")
	} else {
		typeName = strings.ReplaceAll(typeName, ".", "")
	}

	return dag.TypeDef().WithObject(typeName), nil
}

func fieldToFunction(ctx context.Context, dag *dagger.Client,
	field value.ObjectSchemaField) (*dagger.Function, bool, error) {
	if !isPublic(field.Key) ||
		field.Schema == nil ||
		field.Schema.TargetKind() != value.FuncKind ||
		field.Schema.FuncSchema == nil ||
		field.Optional ||
		field.Match {
		return nil, false, nil
	}

	return funcToFunction(ctx, dag, field.Key, field.Description, field.Schema.FuncSchema)
}

func funcToFunction(ctx context.Context, dag *dagger.Client,
	funcName, funcDescription string, funcSchema *value.FuncSchema) (*dagger.Function, bool, error) {
	returnSchema, ok, err := funcSchema.Returns(ctx)
	if err != nil {
		return nil, false, err
	} else if !ok {
		returnSchema = nil
	}

	returnType, err := toTypeDef(ctx, dag, returnSchema)
	if err != nil {
		return nil, false, err
	}

	dagFunc := dag.Function(funcName, returnType).WithDescription(funcDescription)

	for _, arg := range funcSchema.Args {
		if arg.Match {
			continue
		}
		argType, err := toTypeDef(ctx, dag, arg.Schema)
		if err != nil {
			return nil, false, err
		}

		var (
			description  = arg.Description
			defaultValue dagger.JSON
		)

		if arg.Optional {
			argType = argType.WithOptional(arg.Optional)
		}

		def, ok, err := value.DefaultValue(arg.Schema)
		if err != nil {
			return nil, false, err
		}
		if ok {
			defValue, err := json.Marshal(def)
			if err != nil {
				return nil, false, err
			}
			defaultValue = dagger.JSON(defValue)
		}
		dagFunc = dagFunc.WithArg(arg.Key, argType, dagger.FunctionWithArgOpts{
			Description:  description,
			DefaultValue: defaultValue,
		})
	}

	return dagFunc, true, nil
}

func isValidObjectField(field value.ObjectSchemaField) bool {
	return isPublic(field.Key) &&
		field.Schema != nil &&
		field.Schema.TargetKind() == value.ObjectKind &&
		field.Schema.Object != nil &&
		!field.Optional &&
		!field.Match
}

func isValidFuncField(field value.ObjectSchemaField) bool {
	return isPublic(field.Key) &&
		field.Schema != nil &&
		field.Schema.TargetKind() == value.FuncKind &&
		field.Schema.FuncSchema != nil &&
		!field.Optional &&
		!field.Match
}

func getOrCreateObject(dag *dagger.Client, name, description string,
	types map[string]*dagger.TypeDef) *dagger.TypeDef {
	ts, ok := types[name]
	if ok {
		return ts
	}

	obj := dag.TypeDef().WithObject(name, dagger.TypeDefWithObjectOpts{
		Description: description,
	})
	types[name] = obj
	return obj
}

func typeSchemaToObject(ctx context.Context, dag *dagger.Client, types map[string]*dagger.TypeDef, objName string, ts *value.TypeSchema) error {
	obj := getOrCreateObject(dag, objName, ts.Object.Description, types)

	for _, field := range ts.Object.Fields {
		if isValidFuncField(field) {
			f, ok, err := fieldToFunction(ctx, dag, field)
			if err != nil {
				return err
			} else if !ok {
				continue
			}
			obj = obj.WithFunction(f)
			types[objName] = obj
			continue
		}

		if value.IsSimpleKind(field.Schema.TargetKind()) && len(field.Key) > 0 && strings.ToLower(field.Key[:1]) == field.Key[:1] {
			t, err := toTypeDef(ctx, dag, field.Schema)
			if err != nil {
				return err
			}
			obj = obj.WithField(field.Key, t)
			types[objName] = obj

			obj = obj.WithFunction(dag.Function("With"+strings.ToUpper(field.Key[:1])+field.Key[1:], obj).
				WithDescription(fmt.Sprintf("Set %s field", field.Key)).
				WithArg(field.Key, t))
			types[objName] = obj
			continue
		}
	}

	return nil
}

func Register(ctx context.Context, dag *dagger.Client, moduleName string, v value.Value) (*dagger.Module, error) {
	module := dag.CurrentModule()

	entries, err := value.Entries(v)
	if err != nil {
		return nil, err
	}
	var types = map[string]*dagger.TypeDef{}

	for _, entry := range entries {
		ts, ok := entry.Value.(*value.TypeSchema)
		if ok && isPublic(entry.Key) && ts.Object != nil {
			if err := typeSchemaToObject(ctx, dag, types, entry.Key, ts); err != nil {
				return nil, err
			}
			continue
		}

		if f, ok := entry.Value.(*eval.Function); ok {
			daggerFunc, ok, err := funcToFunction(ctx, dag, entry.Key, "", f.FuncSchema)
			if err != nil {
				return nil, err
			} else if !ok {
				continue
			}

			obj := getOrCreateObject(dag, moduleName, "", types)
			obj = obj.WithFunction(daggerFunc)
			types[moduleName] = obj
			continue
		}
	}

	for _, name := range sortedMapKeys(types) {
		module = module.WithObject(types[name])
	}

	return module, nil
}

func sortedMapKeys[T cmp.Ordered, V any](v map[T]V) (result []T) {
	for k := range v {
		result = append(result, k)
	}
	slices.Sort(result)
	return
}
