module github.com/ibuildthecloud/dagamole

go 1.21.2

// Without this jetbrains complains but it does seem to really be needed
replace github.com/dagger/dagger/internal/mage => ../../dagger/internal/mage

require (
	dagger.io/dagger v0.9.3
	github.com/acorn-io/aml v0.0.0-20231106071231-26ca3a01201a
	github.com/dagger/dagger v0.9.3
)

require github.com/kr/text v0.2.0 // indirect

require (
	github.com/99designs/gqlgen v0.17.34 // indirect
	github.com/Khan/genqlient v0.6.0 // indirect
	github.com/adrg/xdg v0.4.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/buildkit v0.13.0-beta1.0.20231011161957-86e25b3ad8c2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/vektah/gqlparser/v2 v2.5.6 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/sync v0.4.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
