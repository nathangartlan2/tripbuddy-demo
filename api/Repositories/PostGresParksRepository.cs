using System;
using api.Models;
using Npgsql;

namespace api.Repositories;

public class PostGresParksRepository : IParksRepository
{
    string _connectionString;

    public PostGresParksRepository(string connectionString)
    {
        _connectionString = connectionString;
    }

    public async Task<IResult> CreateParkAsync(Park park)
    {
        throw new NotImplementedException();

    }

    public Task<IResult> GetParkAsync(string id)
    {
        throw new NotImplementedException();
    }

    public async Task<IResult> GetParksAsync()
    {
        List<Park> parks = new();

        // Open connection
        await using var conn = new NpgsqlConnection(_connectionString);
        await conn.OpenAsync();

        var sql = @"
          SELECT 
              p.id, 
              p.name, 
              p.state_code, 
              p.latitude, 
              p.longitude,
              COALESCE(
                  json_agg(
                      json_build_object('Name', a.name, 'description', a.description)
                  ) FILTER (WHERE a.id IS NOT NULL),
                  '[]'
              ) as activities
          FROM parks p
          LEFT JOIN activities a ON p.id = a.park_id
          GROUP BY p.id, p.name, p.state_code, p.latitude, p.longitude";

        await using var cmd = new NpgsqlCommand(sql, conn);
        await using var reader = await cmd.ExecuteReaderAsync();

        while (await reader.ReadAsync())
        {
            parks.Add(new Park
            {
                Id = reader.GetInt32(0).ToString(),
                Name = reader.GetString(1),
                StateCode = reader.GetString(2),
                Latitude = reader.GetFloat(3),
                Longitude = reader.GetFloat(4),
                Activities = System.Text.Json.JsonSerializer.Deserialize<Activity[]>(reader.GetString(5)) ??
    Array.Empty<Activity>()
            });
        }


        return Results.Ok(parks);
    }
}
