package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var (
	nodeEngineBuildpack                     string
	nodeEngineOfflineBuildpack              string
	icuBuildpack                            string
	icuOfflineBuildpack                     string
	dotnetCoreAspNetRuntimeBuildpack        string
	dotnetCoreAspNetRuntimeOfflineBuildpack string
	dotnetCoreSDKBuildpack                  string
	dotnetCoreSDKOfflineBuildpack           string
	dotnetExecuteBuildpack                  string
	vsdbgBuildpack                          string
	buildpack                               string
	offlineBuildpack                        string
	buildpackInfo                           struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}
	config struct {
		NodeEngine              string `json:"node-engine"`
		ICU                     string `json:"icu"`
		DotnetCoreAspNetRuntime string `json:"dotnet-core-aspnet-runtime"`
		DotnetCoreSDK           string `json:"dotnet-core-sdk"`
		DotnetExecute           string `json:"dotnet-execute"`
		Vsdbg                   string `json:"vsdbg"`
	}
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.NewDecoder(file).Decode(&buildpackInfo)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	buildpack, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	offlineBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	nodeEngineBuildpack, err = buildpackStore.Get.
		Execute(config.NodeEngine)
	Expect(err).NotTo(HaveOccurred())

	nodeEngineOfflineBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion("1.2.3").
		Execute(config.NodeEngine)
	Expect(err).NotTo(HaveOccurred())

	icuBuildpack, err = buildpackStore.Get.
		Execute(config.ICU)
	Expect(err).NotTo(HaveOccurred())

	icuOfflineBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion("1.2.3").
		Execute(config.ICU)
	Expect(err).NotTo(HaveOccurred())

	dotnetCoreAspNetRuntimeBuildpack, err = buildpackStore.Get.
		Execute(config.DotnetCoreAspNetRuntime)
	Expect(err).NotTo(HaveOccurred())

	dotnetCoreAspNetRuntimeOfflineBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion("1.2.3").
		Execute(config.DotnetCoreAspNetRuntime)
	Expect(err).NotTo(HaveOccurred())

	dotnetCoreSDKBuildpack, err = buildpackStore.Get.
		Execute(config.DotnetCoreSDK)
	Expect(err).NotTo(HaveOccurred())

	dotnetCoreSDKOfflineBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion("1.2.3").
		Execute(config.DotnetCoreSDK)
	Expect(err).NotTo(HaveOccurred())

	dotnetExecuteBuildpack, err = buildpackStore.Get.
		Execute(config.DotnetExecute)
	Expect(err).NotTo(HaveOccurred())

	vsdbgBuildpack, err = buildpackStore.Get.
		Execute(config.Vsdbg)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(30 * time.Second)
	format.MaxLength = 0

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Console", testConsole)
	suite("DefaultApps", testDefaultApps)
	suite("FSharp", testFSharp)
	suite("MatchDirAndAppName", testMatchDirAndAppName)
	suite("MultipleProject", testMultipleProject)
	suite("Nuget", testNugetConfig)
	suite("Offline", testOffline)
	suite("OutputSlicing", testOutputSlicing)
	suite("SourceRemoval", testSourceRemoval)
	suite("VisualBasic", testVisualBasic)
	suite.Run(t)
}
