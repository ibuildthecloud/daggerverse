package bootstrap

import (
	"context"
	// so that embedding will work
	_ "embed"
	"encoding/json"
	"os"

	"dagger.io/dagger"
	"github.com/acorn-io/aml"
	"github.com/acorn-io/aml/pkg/value"
	"github.com/ibuildthecloud/dagamole/pkg/dagvalues"
	"github.com/ibuildthecloud/dagamole/pkg/module"
)

const (
	Filename = "dag.aml"
)

var (
	Dagger value.Value
)

func unmarshal(ctx context.Context, globals map[string]any, data []byte, source string, out any) error {
	return aml.Unmarshal(data, out, aml.DecoderOption{
		Context:    ctx,
		SourceName: source,
		Globals:    globals,
	})
}

func Main(ctx context.Context) error {
	var (
		dagaml  value.Value
		globals = map[string]any{}
	)

	dagAML, err := aml.ReadFile(Filename)
	if err != nil {
		return err
	}

	c, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}

	globals["dag"], err = dagvalues.Dagger(ctx, c)
	if err != nil {
		return err
	}

	if err := unmarshal(ctx, globals, dagAML, Filename, &dagaml); err != nil {
		return err
	}

	if ok, err := module.IsModule(ctx, c); err != nil {
		return err
	} else if ok {
		return module.Invoke(ctx, c, dagaml)
	}

	nv, ok, err := value.NativeValue(dagaml)
	if err != nil {
		return err
	}

	if ok {
		return printOutput(nv)
	}

	return nil
}

func printOutput(out any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
