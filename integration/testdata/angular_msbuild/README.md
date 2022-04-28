This app was generated with the .NET Core CLI version 6.0.200 by running:
```
dotnet new angular -o angular_msbuild
```

Then, in `angular_msbuild.csproj`, replace the lines:
```xml
    <!-- As part of publishing, ensure the JS resources are freshly built in production mode -->
    <Exec WorkingDirectory="$(SpaRoot)" Command="npm install" />
```
with the lines:
```xml
    <!-- As part of publishing, ensure the JS resources are freshly built in production mode -->
    <Exec WorkingDirectory="$(SpaRoot)" Command="npm --dev install" />
```
This allows the `ng` CLI to be installed during the build.

This fixture is used to test the case that the Node Engine buildpack works with the .NET
Core language family to server Javascript content.
