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
	logger := scribe.NewLogger(os.Stdout)

	packit.Run(
		dotnetpublish.Detect(
			dotnetpublish.NewDotnetProjectFileParser(),
			bpYMLParser,
		),
		dotnetpublish.Build(
			dotnetpublish.NewDotnetRootManager(),
			dotnetpublish.NewDotnetPublishProcess(
				pexec.NewExecutable("dotnet"),
				logger,
				chronos.DefaultClock,
			),
			bpYMLParser,
			chronos.DefaultClock,
			logger,
		),
	)
}
