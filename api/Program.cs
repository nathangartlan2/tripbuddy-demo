using api.Models;
using api.Repositories;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.StaticAssets;

var builder = WebApplication.CreateBuilder(args);
var app = builder.Build();

string postGresConnection = @"Host=localhost;Port=5432;Database=tripbuddy;Username=tripbuddy_user;Password=tripbuddy_pass";

IParksRepository parkRepo = new PostGresParksRepository(postGresConnection);

app.MapGet("/", () => "TripBuddy API up and running");

app.MapGet("/parks", async () => await parkRepo.GetParksAsync());

app.MapGet("/park/{id}", async (string id) => await parkRepo.GetParkAsync(id));

app.MapGet("/park/search", async (
      [FromQuery] double? latitude, [FromQuery] double? longitude,
      [FromQuery] string activity,
      [FromQuery] double radiusKm = 50) =>
{
    if (latitude.HasValue && longitude.HasValue)
    {
        return Results.Ok(await parkRepo.SearchGeographic(
            latitude.Value,
            longitude.Value,
            activity ?? "",
            radiusKm
        ));
    }

    return Results.Ok(await parkRepo.GetParksAsync());
});

app.Run();
