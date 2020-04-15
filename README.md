# Dotnet Build Cloud Native Buildpack

The Dotnet Core Build CNB requires a set of buildpacks and then compiles the application that
it has been given.

## Integration

The Dotnet Core Build CNB provides build dependency. The build dependency can required
by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the Dotnet Core Build dependency is "build". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "build"

  # Note: The version field is unsupported as there is no version for a set of
  # build.

  # The Dotnet Core Build CNB does not support non-required metadata options.
```

## Usage
To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's source using GOOS=linux by default. You can supply another value as the first argument to package.sh.
