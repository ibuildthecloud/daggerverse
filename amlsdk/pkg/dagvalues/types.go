package dagvalues

import (
	"fmt"

	"github.com/acorn-io/aml/pkg/value"
	"github.com/dagger/dagger/cmd/codegen/introspection"
)

func isScalar(t *introspection.TypeRef) bool {
	if t.IsScalar() {
		return true
	}
	if t.IsList() && realType(t).OfType.IsScalar() {
		return true
	}
	return false
}

func realType(t *introspection.TypeRef) *introspection.TypeRef {
	if t.IsOptional() {
		return t
	}
	return t.OfType
}

//func fieldsToArgValues(f *introspection.Type, args []value.CallArgument) ([]value.Value, error) {
//	var argValues []value.Value
//
//	for i, arg := range args {
//		if arg.Positional {
//			optCount := 0
//			optionalName := ""
//			for _, field := range f.Fields {
//				if !field.TypeRef.IsOptional() {
//					if optCount == i {
//						optionalName = field.Name
//						break
//					} else {
//						optCount++
//					}
//				}
//			}
//			if optionalName == "" {
//				return nil, fmt.Errorf("failed to find required argument to assign value %v to", arg)
//			}
//			argValues = append(argValues, value.NewObject(map[string]any{
//				optionalName: arg.Value,
//			}))
//		} else {
//			argValues = append(argValues, arg.Value)
//		}
//	}
//
//	return argValues, nil
//}

func fieldArgsToValue(f *introspection.Field, args []value.CallArgument) (value.Value, error) {
	var argValues []value.Value

	for i, arg := range args {
		if arg.Positional {
			optCount := 0
			optionalName := ""
			for _, field := range f.Args {
				if !field.TypeRef.IsOptional() {
					if optCount == i {
						optionalName = field.Name
						break
					} else {
						optCount++
					}
				}
			}
			if optionalName == "" {
				return nil, fmt.Errorf("failed to find required argument to assign value %v to", arg)
			}
			argValues = append(argValues, value.NewObject(map[string]any{
				optionalName: arg.Value,
			}))
		} else {
			argValues = append(argValues, arg.Value)
		}
	}

	v, err := value.Merge(argValues...)
	if err != nil {
		return nil, err
	} else if v == nil {
		return value.NewObject(nil), nil
	}
	return v, nil
}
