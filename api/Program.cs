using api.Models;
using api.Repositories;

var builder = WebApplication.CreateBuilder(args);
var app = builder.Build();

IParksRepository parkRepo = new FileParkRepository();

app.MapGet("/", () => "Hello World!");

app.MapGet("/parks", async () => await parkRepo.GetParksAsync());

app.MapGet("/park/{id}", async (string id) => await parkRepo.GetParkAsync(id));

app.Run();
