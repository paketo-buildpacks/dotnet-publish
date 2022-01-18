package main

import (
	"os"

	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func main() {
	bpYMLParser := dotnetpublish.NewDotnetBuildpackYMLParser()
	configParser := dotnetpublish.NewCommandConfigurationParser()
	logger := scribe.NewLogger(os.Stdout)

	packit.Run(
		dotnetpublish.Detect(
			dotnetpublish.NewProjectFileParser(),
			bpYMLParser,
		),
		dotnetpublish.Build(
			dotnetpublish.NewDotnetSourceRemover(),
			dotnetpublish.NewDotnetPublishProcess(
				pexec.NewExecutable("dotnet"),
				logger,
				chronos.DefaultClock,
			),
			bpYMLParser,
			configParser,
			chronos.DefaultClock,
			logger,
		),
	)
}
