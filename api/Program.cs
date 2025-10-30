using api.Models;
using api.Repositories;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.StaticAssets;

var builder = WebApplication.CreateBuilder(args);

// Add Swagger/OpenAPI services
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

var app = builder.Build();

// Configure Swagger middleware
app.UseSwagger();
app.UseSwaggerUI();

// Get connection string from configuration (environment variable or appsettings.json)
// In Docker: uses ConnectionStrings__PostgreSQL environment variable
// Locally: falls back to default localhost connection
string postGresConnection = builder.Configuration.GetConnectionString("PostgreSQL")
    ?? @"Host=localhost;Port=5432;Database=tripbuddy;Username=tripbuddy_user;Password=tripbuddy_pass";

IParksRepository parkRepo = new PostGresParksRepository(postGresConnection);

app.MapGet("/parks", async () => await parkRepo.GetParksAsync());

app.MapGet("/park/{id}", async (string id) => await parkRepo.GetParkAsync(id));

app.MapPost("/park", async ([FromBody] Park park) => await parkRepo.CreateParkAsync(park));

app.MapPut("/park/{parkCode}", async (string parkCode, [FromBody] Park park) => await parkRepo.UpdateParkAsync(parkCode, park));

app.MapDelete("/park/{id}", async (string id) => await parkRepo.DeleteParkAsync(id));

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
