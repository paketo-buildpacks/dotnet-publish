package dotnetpublish_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDotnetPublish(t *testing.T) {
	suite := spec.New("dotnet-publish", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite("DotnetPublishProcess", testDotnetPublishProcess)
	suite("DotnetSourceRemover", testDotnetSourceRemover)
	suite("ProjectFileParser", testProjectFileParser)
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite.Run(t)
}
