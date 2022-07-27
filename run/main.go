package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Netflix/go-env"
	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
)

func main() {
	var config dotnetpublish.Configuration
	_, err := env.UnmarshalFromEnviron(&config)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse build configuration: %w", err))
	}

	bpYMLParser := dotnetpublish.NewDotnetBuildpackYMLParser()
	logger := scribe.NewEmitter(os.Stdout).WithLevel(config.LogLevel)
	bindingResolver := servicebindings.NewResolver()
	symlinker := dotnetpublish.NewSymlinker()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	packit.Run(
		dotnetpublish.Detect(
			config,
			dotnetpublish.NewProjectFileParser(),
			bpYMLParser,
		),
		dotnetpublish.Build(
			config,
			dotnetpublish.NewDotnetSourceRemover(),
			bindingResolver,
			homeDir,
			symlinker,
			dotnetpublish.NewDotnetPublishProcess(
				pexec.NewExecutable("dotnet"),
				logger,
				chronos.DefaultClock,
			),
			dotnetpublish.NewOutputSlicer(),
			bpYMLParser,
			chronos.DefaultClock,
			logger,
		),
	)
}
