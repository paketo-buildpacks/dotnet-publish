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
	nodeEngineBuildpack        string
	icuBuildpack               string
	dotnetCoreRuntimeBuildpack string
	dotnetCoreAspNetBuildpack  string
	dotnetCoreSDKBuildpack     string
	dotnetExecuteBuildpack     string
	buildpack                  string
	offlineBuildpack           string
	buildpackInfo              struct {
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

	icuBuildpack, err = buildpackStore.Get.
		Execute(config.ICU)
	Expect(err).NotTo(HaveOccurred())

	dotnetCoreRuntimeBuildpack, err = buildpackStore.Get.
		Execute(config.DotnetCoreRuntime)
	Expect(err).NotTo(HaveOccurred())

	dotnetCoreAspNetBuildpack, err = buildpackStore.Get.
		Execute(config.DotnetCoreAspNet)
	Expect(err).NotTo(HaveOccurred())

	dotnetCoreSDKBuildpack, err = buildpackStore.Get.
		Execute(config.DotnetCoreSDK)
	Expect(err).NotTo(HaveOccurred())

	dotnetExecuteBuildpack, err = buildpackStore.Get.
		Execute(config.DotnetExecute)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(30 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Sequential())
	suite("Default", testDefault)
	suite("SourceRemoval", testSourceRemoval)
	suite.Run(t)
}
