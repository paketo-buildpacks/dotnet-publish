package main

import (
	"encoding/xml"
	"fmt"
	"github.com/cloudfoundry/dotnet-core-build-cnb/publish"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
)

type Proj struct {
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

	projFile, err := getProjFile(context.Application.Root)
	if err != nil{
		return context.Fail(), err
	}

	projObj, err := parseProj(projFile)
	if err != nil{
		return context.Fail(), err
	}

	version := resolveVersion(projObj)

	if detectASPNet(projObj){
		plan.Requires = []buildplan.Required{{
			Name:	  publish.Publish,
			Metadata: buildplan.Metadata{"build": true},
		},{
			Name:     "dotnet-sdk",
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch":true },
		},{
			Name:     "dotnet-runtime",
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		},{
			Name:     "dotnet-aspnet",
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		}}
	} else {
		plan.Requires = []buildplan.Required{{
			Name:     publish.Publish,
			Metadata: buildplan.Metadata{"build": true},
		}, {
			Name:     "dotnet-sdk",
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		}, {
			Name:     "dotnet-runtime",
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		}}
	}

	return context.Pass(plan)
}

func getProjFile(appRoot string) (string, error){
	fileName, err := filepath.Glob(filepath.Join(appRoot,"*.?sproj"))
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
	matches := regexp.MustCompile(`netcoreapp(.*)`).FindStringSubmatch(projObj.PropertyGroup.TargetFramework)

	return fmt.Sprintf("%s.0", matches[1])
}

func detectASPNet(projObj Proj) bool {
	for _, ig := range projObj.ItemGroups {
		for _, pr := range ig.PackageReferences {
			if pr.Include == "Microsoft.AspNetCore.App" || pr.Include == "Microsoft.AspNetCore.All" {
				return true
			}
		}
	}
	return false
}

