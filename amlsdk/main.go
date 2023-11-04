package main

import "path/filepath"

type Amlsdk struct{}

func (m *Amlsdk) ModuleRuntime(modSource *Directory, subPath string, introspectionJson string) *Container {
	return m.Base().
		WithMountedDirectory("/src", modSource).
		WithWorkdir(filepath.Join("/src", subPath))
}

func (m *Amlsdk) Codegen(modSource *Directory, subPath string, introspectionJson string) *GeneratedCode {
	return dag.GeneratedCode(dag.Directory())
}

func (m *Amlsdk) Base() *Container {
	bin := dag.Container().
		From("cgr.dev/chainguard/go").
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-cache")).
		WithMountedCache("/root/go/pkg/mod", dag.CacheVolume("go-mod-cache")).
		WithMountedDirectory("/src", dag.Host().Directory(".")).
		WithWorkdir("/src").
		WithExec([]string{"build", "-o", "/src/bin/dagamole", "./cmd/dagamole"}).
		File("/src/bin/dagamole")

	return dag.Container().
		From("cgr.dev/chainguard/go").
		WithFile("/usr/local/bin/dagamole", bin).
		WithEntrypoint([]string{"/usr/local/bin/dagamole"}).
		WithDefaultArgs()
}
