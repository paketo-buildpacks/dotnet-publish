package dotnetpublish_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDotnetPublish(t *testing.T) {
	suite := spec.New("dotnet-publish", spec.Report(report.Terminal{}), spec.Sequential())
	suite("Build", testBuild)
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("CommandConfigurationParser", testCommandConfigurationParser, spec.Sequential())
	suite("Detect", testDetect)
	suite("DotnetProcess", testDotnetProcess)
	suite("DotnetSourceRemover", testDotnetSourceRemover)
	suite("ProjectFileParser", testProjectFileParser)
	suite.Run(t)
}
