package dotnetpublish

import (
	"fmt"
	"path/filepath"

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

func Detect(config Configuration, parser ProjectParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
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

		requirements = append(requirements, packit.BuildPlanRequirement{
			Name: "icu",
			Metadata: BuildPlanMetadata{
				Build: true,
			},
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
