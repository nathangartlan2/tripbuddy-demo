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

    public async Task<IResult> GetParkAsync(string id)
    {
        Park park;
        await using var conn = new NpgsqlConnection(_connectionString);
        await conn.OpenAsync();

        var sql = @"SELECT
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
		WHERE p.id = 1
		GROUP BY p.id, p.name, p.state_code, p.latitude, p.longitude;";

        await using var cmd = new NpgsqlCommand(sql, conn);
        await using var reader = await cmd.ExecuteReaderAsync();

        while (await reader.ReadAsync())
        {
            park = new Park
            {
                Id = reader.GetInt32(0).ToString(),
                Name = reader.GetString(1),
                StateCode = reader.GetString(2),
                Latitude = reader.GetFloat(3),
                Longitude = reader.GetFloat(4),
                Activities = System.Text.Json.JsonSerializer.Deserialize<Activity[]>(reader.GetString(5)) ??
                 Array.Empty<Activity>()
            };

            return Results.Ok(park);
        }


        return Results.NotFound();
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

    public async Task<IResult> SearchGeographic(double latitude, double longitude, string activity, double radiusKm)
    {

        List<Park> parks = new();

        await using var conn = new NpgsqlConnection(_connectionString);
        await conn.OpenAsync();

        var sql = @$"SELECT
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
                        ) as activities,
            ST_Distance(location, ST_MakePoint({longitude}, {latitude})::geography) / 1000 AS distance_km
            FROM parks p
            LEFT JOIN activities a ON p.id = a.park_id
            WHERE
            to_tsvector('english', a.name) @@ to_tsquery('english', '{activity}')
            AND ST_DWithin(location, ST_MakePoint({longitude}, {latitude})::geography, {radiusKm * 1000})
            GROUP BY p.id, p.name, p.state_code, p.latitude, p.longitude, distance_km
            ORDER BY distance_km;";

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
