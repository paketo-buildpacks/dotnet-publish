package dotnetpublish

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/dotnet-publish/internal"
	"github.com/paketo-buildpacks/packit/v2"
)

type namedSlices map[string]map[string]string

func addPath(slices namedSlices, name, element string) namedSlices {
	if _, ok := slices[name]; !ok {
		slices[name] = map[string]string{}
	}
	slices[name][element] = ""
	return slices
}

type OutputSlicer struct{}

func NewOutputSlicer() OutputSlicer {
	return OutputSlicer{}
}

func (s OutputSlicer) Slice(assetsFile string) (pkgs, earlyPkgs, projects packit.Slice, err error) {
	contents, err := os.Open(assetsFile)
	if err != nil {
		return packit.Slice{}, packit.Slice{}, packit.Slice{}, fmt.Errorf("opening assets file to identify output slices: %w", err)
	}
	defer contents.Close()

	var assets internal.ProjectAssetsJSON
	dec := json.NewDecoder(contents)
	err = dec.Decode(&assets)
	if err != nil {
		return packit.Slice{}, packit.Slice{}, packit.Slice{}, fmt.Errorf("decoding JSON to identify output slices: %w", err)
	}

	slices := namedSlices{}

	for _, target := range assets.Targets {
		for _, dep := range target.Dependencies {
			if dep.Type != "package" && dep.Type != "project" {
				continue
			}
			if dep.Type == "package" {
				version := strings.Split(dep.Name, "/")[1] // back half of dep name is version
				if strings.Contains(version, "-") {        // version with dash is a release candidate or beta
					dep.Type = "early-package"
				}
			}
			file := filepath.Base(string(dep.Runtime))
			if file != "" && file != "_._" {
				slices = addPath(slices, dep.Type, file)
			}

			for _, rt := range dep.RuntimeTargets {
				slices = addPath(slices, dep.Type, filepath.Base(rt.FileName))
			}
		}
		// TODO: What if same dependency gets filed under two different types?
	}

	for name, paths := range slices {
		var slicePaths *[]string
		switch name {
		case "package":
			slicePaths = &pkgs.Paths
		case "early-package":
			slicePaths = &earlyPkgs.Paths
		case "project":
			slicePaths = &projects.Paths
		}

		for path := range paths {
			*slicePaths = append(*slicePaths, path)
		}
	}
	return pkgs, earlyPkgs, projects, nil
}
