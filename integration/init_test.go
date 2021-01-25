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
)

var (
	nodeEngineBuildpack               string
	nodeEngineOfflineBuildpack        string
	icuBuildpack                      string
	icuOfflineBuildpack               string
	dotnetCoreRuntimeBuildpack        string
	dotnetCoreRuntimeOfflineBuildpack string
	dotnetCoreAspNetBuildpack         string
	dotnetCoreAspNetOfflineBuildpack  string
	dotnetCoreSDKBuildpack            string
	dotnetCoreSDKOfflineBuildpack     string
	dotnetExecuteBuildpack            string
	buildpack                         string
	offlineBuildpack                  string
	buildpackInfo                     struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}
	config struct {
		NodeEngine        string `json:"node-engine"`
		ICU               string `json:"icu"`
		DotnetCoreRuntime string `json:"dotnet-core-runtime"`
		DotnetCoreAspNet  string `json:"dotnet-core-aspnet"`
		DotnetCoreSDK     string `json:"dotnet-core-sdk"`
		DotnetExecute     string `json:"dotnet-execute"`
	}
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.DecodeReader(file, &buildpackInfo)
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

	dotnetCoreRuntimeBuildpack, err = buildpackStore.Get.
		Execute(config.DotnetCoreRuntime)
	Expect(err).NotTo(HaveOccurred())

	dotnetCoreRuntimeOfflineBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion("1.2.3").
		Execute(config.DotnetCoreRuntime)
	Expect(err).NotTo(HaveOccurred())

	dotnetCoreAspNetBuildpack, err = buildpackStore.Get.
		Execute(config.DotnetCoreAspNet)
	Expect(err).NotTo(HaveOccurred())

	dotnetCoreAspNetOfflineBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion("1.2.3").
		Execute(config.DotnetCoreAspNet)
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

	SetDefaultEventuallyTimeout(30 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("21Apps", test21Apps)
	suite("31Apps", test31Apps)
	suite("ASPNet", testAspnetSource)
	suite("Console", testConsole)
	suite("Default", testDefault)
	suite("SourceRemoval", testSourceRemoval)
	suite("FSharp", testFSharp)
	suite("Kestrel", testKestrel)
	suite("MultipleProject", testMultipleProject)
	suite("SelfContained", testSelfContained)
	suite("Versioning", testVersioning)
	suite.Run(t)
}
