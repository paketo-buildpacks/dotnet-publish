using console_app;

var builder = WebApplication.CreateBuilder(args);
var app = builder.Build();

app.MapGet("/", () => ProjectTwo.GetAString());

app.Run();
