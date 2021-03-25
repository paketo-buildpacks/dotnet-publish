# Dotnet Publish Cloud Native Buildpack

The Dotnet Publish CNB requires a set of buildpacks and then compiles the application that
it has been given.

## Integration

The Dotnet Publish CNB provides build dependency. The build dependency can required
by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the Dotnet Publish dependency is "build". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "build"

  # Note: The version field is unsupported as there is no version for a set of
  # build.

  # The Dotnet Publish CNB does not support non-required metadata options.
```

## Usage
To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## Specifying a project path

To specify a project subdirectory to be used as the root of the app, please use
the BP_DOTNET_PROJECT_PATH environment variable at build time either directly
(e.g. pack build my-app --env BP_DOTNET_PROJECT_PATH=./src/my-app) or through a
project.toml file.
