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

// Initialize the parks repository using the factory
IParksRepository parkRepo = ParkRepoFactory(builder.Configuration);

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

static IParksRepository ParkRepoFactory(ConfigurationManager config)
{
    // Get repository type from configuration (default to PostgreSQL)
    string repositoryType = config["RepositoryType"] ?? "PostgreSQL";

    return repositoryType.ToLower() switch
    {
        "file" => new FileParkRepository(),

        "postgresql" or "postgres" => CreatePostgreSqlRepository(config),

        _ => throw new InvalidOperationException(
            $"Unknown repository type: '{repositoryType}'. Valid options: 'File', 'PostgreSQL'")
    };
}

static PostGresParksRepository CreatePostgreSqlRepository(ConfigurationManager config)
{
    // Get connection string from configuration
    // Priority order:
    // 1. DATABASE_URL (Fly.io automatically sets this when Postgres is attached)
    // 2. ConnectionStrings:PostgreSQL (local dev, Docker, or manual Fly.io secret)
    string connectionString = config["DATABASE_URL"]
        ?? config.GetConnectionString("PostgreSQL")
        ?? throw new InvalidOperationException(
            "PostgreSQL connection string is not configured. " +
            "Set 'DATABASE_URL' environment variable (Fly.io) or 'ConnectionStrings:PostgreSQL' in configuration.");

    return new PostGresParksRepository(connectionString);
}