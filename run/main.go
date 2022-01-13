package main

import (
	"log"
	"os"

	dotnetpublish "github.com/paketo-buildpacks/dotnet-publish"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/paketo-buildpacks/packit/servicebindings"
)

func main() {
	bpYMLParser := dotnetpublish.NewDotnetBuildpackYMLParser()
	configParser := dotnetpublish.NewCommandConfigurationParser()
	logger := scribe.NewLogger(os.Stdout)
	bindingResolver := servicebindings.NewResolver()
	symlinker := dotnetpublish.NewSymlinker()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	packit.Run(
		dotnetpublish.Detect(
			dotnetpublish.NewProjectFileParser(),
			bpYMLParser,
		),
		dotnetpublish.Build(
			dotnetpublish.NewDotnetSourceRemover(),
			bindingResolver,
			homeDir,
			symlinker,
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
