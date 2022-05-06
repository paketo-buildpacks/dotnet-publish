package internal_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitInternal(t *testing.T) {
	suite := spec.New("internal", spec.Report(report.Terminal{}))
	suite("Targets", testTargets)
	suite("RuntimeTargets", testRuntimeTargets)
	suite("Runtime", testRuntime)
	suite("Dependencies", testDependencies)
	suite.Run(t)
}
