This app was created by:
1. With .NET Core CLI: `dotnet new sln`
2. Add a console app (`dotnet new console -o console_app`) and web app (`dotnet
   new web -o asp_web_app`). These apps were re-configured so that the web app
   references the `console_app`. Check out the specific code inside of
   `asp_web_app` for how the package references are set up.
3. Add the apps to the solution file:
   `dotnet sln multiple_projects_msbuild.sln add console_app/console_app.csproj`
   `dotnet sln multiple_projects_msbuild.sln add asp_web_app/asp_web_app.csproj`
