package dotnetpublish

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit/v2"
)

type BuildPlanMetadata struct {
	Version       string `toml:"version,omitempty"`
	VersionSource string `toml:"version-source,omitempty"`
	Build         bool   `toml:"build"`
	Launch        bool   `toml:"launch"`
}

//go:generate faux --interface ProjectParser --output fakes/project_parser.go
type ProjectParser interface {
	FindProjectFile(root string) (string, error)
	ParseVersion(path string) (string, error)
	NodeIsRequired(path string) (bool, error)
	NPMIsRequired(path string) (bool, error)
}

//go:generate faux --interface BuildpackYMLParser --output fakes/buildpack_yml_parser.go
type BuildpackYMLParser interface {
	ParseProjectPath(path string) (projectFilePath string, err error)
}

func Detect(config Configuration, parser ProjectParser, buildpackYMLParser BuildpackYMLParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		if config.ProjectPath == "" {
			var err error
			config.ProjectPath, err = buildpackYMLParser.ParseProjectPath(filepath.Join(context.WorkingDir, "buildpack.yml"))
			if err != nil {
				return packit.DetectResult{}, fmt.Errorf("failed to parse buildpack.yml: %w", err)
			}
		}

		projectFilePath, err := parser.FindProjectFile(filepath.Join(context.WorkingDir, config.ProjectPath))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if projectFilePath == "" {
			return packit.DetectResult{}, packit.Fail.WithMessage("no project file found")
		}

		version, err := parser.ParseVersion(projectFilePath)
		if err != nil {
			return packit.DetectResult{}, err
		}

		semver, err := semver.NewVersion(version)
		if err != nil {
			return packit.DetectResult{}, err
		}

		requirements := []packit.BuildPlanRequirement{
			{
				Name: "dotnet-sdk",
				Metadata: BuildPlanMetadata{
					Build:         true,
					Version:       fmt.Sprintf("%d.%d.*", semver.Major(), semver.Minor()),
					VersionSource: filepath.Base(projectFilePath),
				},
			},
		}

		// Determine ICU metadata to include
		// If .NET Core is 3.1 version line, require ICU 70.*
		// See https://forum.manjaro.org/t/dotnet-3-1-builds-fail-after-icu-system-package-updated-to-71-1-1/114232/9 for details
		icuBuildPlanMetadata := BuildPlanMetadata{
			Build: true,
		}

		isDotnet31, err := checkDotnet31(version)
		if err != nil {
			// untested, version will have already failed on line 56 if
			// malformed
			return packit.DetectResult{}, err
		}
		if isDotnet31 {
			icuBuildPlanMetadata.Version = "70.*"
			icuBuildPlanMetadata.VersionSource = "dotnet-31"
		}

		requirements = append(requirements, packit.BuildPlanRequirement{
			Name:     "icu",
			Metadata: icuBuildPlanMetadata,
		})

		nodeReq, err := parser.NodeIsRequired(projectFilePath)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if nodeReq {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "node",
				Metadata: BuildPlanMetadata{
					Build: true,
				},
			})
		}

		npmReq, err := parser.NPMIsRequired(projectFilePath)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if npmReq {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "npm",
				Metadata: BuildPlanMetadata{
					Build: true,
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "dotnet-application"},
				},
				Requires: requirements,
			},
		}, nil
	}
}

func checkDotnet31(version string) (bool, error) {
	match, err := regexp.MatchString(`3\.1\.*`, version)
	if err != nil {
		// untested because regexp pattern is hardcoded
		return false, err
	}

	return match, nil
}
