module github.com/paketo-buildpacks/dotnet-publish

go 1.16

require (
	github.com/BurntSushi/toml v1.2.0
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/Netflix/go-env v0.0.0-20220526054621-78278af1949d
	github.com/docker/docker v20.10.19+incompatible // indirect
	github.com/mattn/go-shellwords v1.0.12
	github.com/moby/term v0.0.0-20220808134915-39b0c02b01ae // indirect
	github.com/onsi/gomega v1.22.1
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/opencontainers/runc v1.1.4 // indirect
	github.com/paketo-buildpacks/occam v0.13.3
	github.com/paketo-buildpacks/packit/v2 v2.6.1
	github.com/sclevine/spec v1.4.0
	golang.org/x/net v0.0.0-20221014081412-f15817d10f9b // indirect
	golang.org/x/sys v0.0.0-20221013171732-95e765b1cc43 // indirect
	golang.org/x/text v0.3.8 // indirect
	google.golang.org/genproto v0.0.0-20221014213838-99cd37c6964a // indirect
	google.golang.org/grpc v1.50.1 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/CycloneDX/cyclonedx-go => github.com/CycloneDX/cyclonedx-go v0.6.0
