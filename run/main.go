package main

import (
	"os"

	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
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
