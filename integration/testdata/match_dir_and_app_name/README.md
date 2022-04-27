This app was generated with the .NET Core CLI version 3.1.413 by running:
```
dotnet new console -o console
mv ./console ./match_dir_and_app_name
mkdir ./console
mv ./match_dir_and_app_name/* ./console
mv ./console ./match_dir_and_app_name/console
```

The resulting app contains a subdirectory with the same name as the compiled
app's entrypoint. This fixture tests an edge case where that naming collision
could cause a build failure.
