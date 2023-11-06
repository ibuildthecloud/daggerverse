package module

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"
	"github.com/acorn-io/aml"
	"github.com/acorn-io/aml/pkg/value"
)

const (
	noModuleError = "not in a module"
)

func Invoke(ctx context.Context, dag *dagger.Client, v value.Value) error {
	if moduleName, ok := isForceModule(); ok {
		_, err := Register(ctx, dag, moduleName, v)
		return err
	}

	fnCall := dag.CurrentFunctionCall()
	parentName, err := fnCall.ParentName(ctx)
	if err != nil {
		return err
	}

	fnName, err := fnCall.Name(ctx)
	if err != nil {
		return err
	}
	parentJson, err := fnCall.Parent(ctx)
	if err != nil {
		return err
	}

	fnArgs, err := fnCall.InputArgs(ctx)
	if err != nil {
		return err
	}

	inputArgs := map[string][]byte{}
	for _, fnArg := range fnArgs {
		argName, err := fnArg.Name(ctx)
		if err != nil {
			return err
		}
		argValue, err := fnArg.Value(ctx)
		if err != nil {
			return err
		}
		inputArgs[argName] = []byte(argValue)
	}

	result, err := invoke(ctx, dag, v, []byte(parentJson), parentName, fnName, inputArgs)
	if err != nil {
		args, _ := json.Marshal(inputArgs)
		return fmt.Errorf("failed to invoke (%s) %s.%s with %s: %v", parentJson, parentName, fnName, args, err)
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	_, err = fnCall.ReturnValue(ctx, dagger.JSON(resultBytes))
	if err != nil {
		return err
	}
	return nil
}

func toParent(data []byte) (*value.Object, error) {
	obj := map[string]any{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	return value.NewObject(obj), nil
}

func toArgs(inputArgs map[string][]byte) (map[string]value.Value, error) {
	result := map[string]value.Value{}
	for _, key := range sortedMapKeys(inputArgs) {
		var val value.Value
		err := aml.Unmarshal(inputArgs[key], &val)
		if err != nil {
			return nil, err
		}
		result[key] = val
	}
	return result, nil
}

func trySetField(schema, obj value.Value, fnName string, argsValues map[string]value.Value) (any, bool, error) {
	if !strings.HasPrefix(fnName, "With") {
		return nil, false, nil
	}

	fieldName := strings.TrimPrefix(fnName, "With")
	if len(fieldName) == 0 {
		return nil, false, nil
	}

	fieldName = strings.ToLower(fieldName[:1]) + fieldName[1:]

	_, ok := argsValues[fieldName]
	if !ok || len(argsValues) != 1 {
		return nil, false, nil
	}

	field, ok, err := value.Lookup(schema, value.NewValue(fieldName))
	if err != nil || !ok {
		return nil, ok, err
	}

	if s, ok := field.(value.Schema); !ok || !value.IsSimpleKind(s.TargetKind()) {
		return nil, false, err
	}

	set, err := value.Add(obj, value.NewValue(argsValues))
	if err != nil {
		return nil, false, err
	}

	return value.NativeValue(set)
}

func invoke(ctx context.Context, dag *dagger.Client, v value.Value, parentJSON []byte, parentName string, fnName string, inputArgs map[string][]byte) (any, error) {
	if parentName == "" {
		moduleName, err := dag.CurrentModule().Name(ctx)
		if err != nil {
			return nil, err
		}
		return Register(ctx, dag, moduleName, v)
	}

	parentData, err := toParent(parentJSON)
	if err != nil {
		return nil, err
	}

	args, err := toArgs(inputArgs)
	if err != nil {
		return nil, err
	}

	return call(ctx, v, parentData, parentName, fnName, args)
}

func call(ctx context.Context, v value.Value, parentData *value.Object, parentName string, fnName string, argValues map[string]value.Value) (any, error) {
	var (
		result   value.Value
		function value.Value
	)

	obj, ok, err := value.Lookup(v, value.NewValue(parentName))
	if err != nil {
		return nil, err
	} else if !ok {
		f, fOk, fErr := value.Lookup(v, value.NewValue(fnName))
		if fErr == nil && fOk {
			function = f
		} else {
			return nil, fmt.Errorf("object now found: %s", parentName)
		}
	} else if ok {
		result, err = value.Validate(ctx, obj, parentData)
		if err != nil {
			return nil, err
		}
	}

	if result != nil {
		function, ok, err = value.Lookup(result, value.NewValue(fnName))
		if err != nil {
			return nil, err
		} else if !ok {
			newResult, ok, newErr := trySetField(obj, result, fnName, argValues)
			if newErr == nil && ok {
				return newResult, nil
			}
			return nil, fmt.Errorf("function not found %s on object %s", fnName, parentName)
		}
	}

	var args []value.CallArgument

	for _, key := range sortedMapKeys(argValues) {
		args = append(args, value.CallArgument{
			Value: value.NewObject(map[string]any{
				key: argValues[key],
			}),
		})
	}

	ret, ok, err := value.Call(ctx, function, args...)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}

	nv, _, err := value.NativeValue(ret)
	return nv, err
}

func isForceModule() (string, bool) {
	env := os.Getenv("DAGAMOLE_MODULE")
	return env, env != ""
}

func IsModule(ctx context.Context, dag *dagger.Client) (bool, error) {
	if _, ok := isForceModule(); ok {
		return ok, nil
	}

	_, err := dag.CurrentModule().Name(ctx)
	if err == nil {
		return true, nil
	} else if strings.Contains(err.Error(), noModuleError) {
		return false, nil
	}
	return false, err
}
