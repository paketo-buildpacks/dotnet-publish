package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudfoundry/dotnet-core-aspnet-cnb/aspnet"
	"github.com/cloudfoundry/dotnet-core-build-cnb/publish"
	"github.com/cloudfoundry/dotnet-core-conf-cnb/utils"
	"github.com/cloudfoundry/dotnet-core-runtime-cnb/runtime"
	"github.com/cloudfoundry/dotnet-core-sdk-cnb/sdk"
	"github.com/cloudfoundry/node-engine-cnb/node"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
)

type Proj struct {
	Sdk           string `xml:"Sdk,attr"`
	PropertyGroup struct {
		TargetFramework         string `xml:"TargetFramework"`
		RuntimeFrameworkVersion string `xml:"RuntimeFrameworkVersion"`
		AssemblyName            string `xml:"AssemblyName"`
	}
	ItemGroups []struct {
		PackageReferences []struct {
			Include string `xml:"Include,attr"`
			Version string `xml:"Version,attr"`
		} `xml:"PackageReference"`
	} `xml:"ItemGroup"`
	Targets []struct {
		Name          string `xml:"Name,attr"`
		BeforeTargets string `xml:"BeforeTargets,attr"`
		AfterTargets  string `xml:"AfterTargets,attr"`
		Exec          []struct {
			Command string `xml:"Command,attr"`
		} `xml:"Exec"`
	} `xml:"Target"`
}

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(100)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	plan := buildplan.Plan{
		Provides: []buildplan.Provided{{Name: publish.Publish}}}

	appRoot, err := utils.GetAppRoot(context.Application.Root)
	if err != nil {
		return context.Fail(), err
	}

	projFile, err := getProjFile(appRoot)
	if err != nil {
		return context.Fail(), err
	}

	projObj, err := parseProj(projFile)
	if err != nil {
		return context.Fail(), err
	}

	version := resolveVersion(projObj)
	splitVersion := strings.Split(version, ".")
	sdkVersion := fmt.Sprintf("%s.%s.0", splitVersion[0], splitVersion[1])

	plan.Requires = []buildplan.Required{{
		Name:     publish.Publish,
		Metadata: buildplan.Metadata{"build": true},
	}, {
		Name:     sdk.DotnetSDK,
		Version:  sdkVersion,
		Metadata: buildplan.Metadata{"build": true, "launch": true},
	}, {
		Name:     runtime.DotnetRuntime,
		Version:  version,
		Metadata: buildplan.Metadata{"build": true, "launch": true},
	}}

	//Parse csproj to find "npm"
	if detectNPM(projObj) {
		plan.Requires = append(plan.Requires, buildplan.Required{
			Name:     node.Dependency,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		})
	}

	if detectASPNet(projObj) {
		plan.Requires = append(plan.Requires, buildplan.Required{
			Name:     aspnet.DotnetAspNet,
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		})
	}

	if context.Stack == "io.buildpacks.stacks.bionic" {
		plan.Requires = append(plan.Requires, buildplan.Required{
			Name:     "icu",
			Metadata: buildplan.Metadata{"build": true},
		})
	}

	return context.Pass(plan)
}

func getProjFile(appRoot string) (string, error) {
	fileName, err := filepath.Glob(filepath.Join(appRoot, "*.?sproj"))
	if err != nil {
		return "", err
	}

	if len(fileName) == 0 {
		return "", fmt.Errorf("no proj file found")
	}

	return fileName[0], nil
}

func parseProj(projPath string) (Proj, error) {

	projBytes, err := ioutil.ReadFile(projPath)
	if err != nil {
		return Proj{}, err
	}

	projObj := Proj{}

	if err := xml.Unmarshal(projBytes, &projObj); err != nil {
		return Proj{}, err
	}

	return projObj, nil
}

func resolveVersion(projObj Proj) string {
	if projObj.PropertyGroup.RuntimeFrameworkVersion == "" {
		matches := regexp.MustCompile(`netcoreapp(.*)`).FindStringSubmatch(projObj.PropertyGroup.TargetFramework)
		return fmt.Sprintf("%s.*", matches[1])
	}

	return projObj.PropertyGroup.RuntimeFrameworkVersion
}

func detectASPNet(projObj Proj) bool {
	// needed to detect steeltoe apps when can ommit Aspnet from the ItemGroup list
	if projObj.Sdk == "Microsoft.NET.Sdk.Web" {
		return true
	}
	for _, ig := range projObj.ItemGroups {
		for _, pr := range ig.PackageReferences {
			if pr.Include == "Microsoft.AspNetCore.App" || pr.Include == "Microsoft.AspNetCore.All" {
				return true
			}
		}
	}
	return false
}

func detectNPM(projObj Proj) bool {

	for _, target := range projObj.Targets {
		for _, ex := range target.Exec {
			command := ex.Command
			if strings.Contains(command, "npm") {
				return true
			}
		}
	}
	return false
}
