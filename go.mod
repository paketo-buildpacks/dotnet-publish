module github.com/paketo-buildpacks/dotnet-publish

go 1.16

require (
	github.com/BurntSushi/toml v1.2.1
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/Microsoft/hcsshim v0.9.5 // indirect
	github.com/Netflix/go-env v0.0.0-20220526054621-78278af1949d
	github.com/anchore/syft v0.62.3 // indirect
	github.com/containerd/containerd v1.6.9 // indirect
	github.com/docker/docker v20.10.21+incompatible // indirect
	github.com/mattn/go-shellwords v1.0.12
	github.com/moby/term v0.0.0-20221105221325-4eb28fa6025c // indirect
	github.com/onsi/gomega v1.24.1
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/opencontainers/runc v1.1.4 // indirect
	github.com/paketo-buildpacks/occam v0.13.3
	github.com/paketo-buildpacks/packit/v2 v2.7.0
	github.com/sclevine/spec v1.4.0
	github.com/testcontainers/testcontainers-go v0.15.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	google.golang.org/genproto v0.0.0-20221027153422-115e99e71e1c // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/CycloneDX/cyclonedx-go => github.com/CycloneDX/cyclonedx-go v0.6.0
