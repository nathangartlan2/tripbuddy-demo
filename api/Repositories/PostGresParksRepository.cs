using System;
using System.Text;
using System.Text.RegularExpressions;
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


    private string ToUrlFriendly(string input)
    {
        if (string.IsNullOrWhiteSpace(input))
            return string.Empty;

        // Convert to lowercase and replace spaces with hyphens
        var result = input.ToLowerInvariant().Trim();
        result = Regex.Replace(result, @"\s+", "-");

        // Remove all characters except alphanumeric, hyphens, and underscores
        result = Regex.Replace(result, @"[^a-z0-9\-_]", "");

        // Replace multiple consecutive hyphens with a single hyphen
        result = Regex.Replace(result, @"-+", "-");

        // Remove leading/trailing hyphens
        result = result.Trim('-');

        return result;
    }
    private string buildNaturalKey(Park park)
    {
        StringBuilder sb = new();
        sb.Append(ToUrlFriendly(park.Name));
        sb.Append("-");
        sb.Append(ToUrlFriendly(park.StateCode));
        return sb.ToString();

    }
    public async Task<IResult> CreateParkAsync(Park park)
    {
        string parkCode = buildNaturalKey(park);

        await using var conn = new NpgsqlConnection(_connectionString);
        await conn.OpenAsync();

        await using var transaction = await conn.BeginTransactionAsync();

        try
        {
            // Insert park with parameterized query
            var insertParkSql = @"
                INSERT INTO parks (name, park_code, park_url, state_code, latitude, longitude)
                VALUES (@name, @parkCode, @parkUrl, @stateCode, @latitude, @longitude)
                RETURNING id";

            int parkId;
            await using (var cmd = new NpgsqlCommand(insertParkSql, conn, transaction))
            {
                cmd.Parameters.AddWithValue("name", park.Name);
                cmd.Parameters.AddWithValue("parkCode", parkCode);
                cmd.Parameters.AddWithValue("parkUrl", DBNull.Value);
                cmd.Parameters.AddWithValue("stateCode", park.StateCode);
                cmd.Parameters.AddWithValue("latitude", park.Latitude);
                cmd.Parameters.AddWithValue("longitude", park.Longitude);

                var result = await cmd.ExecuteScalarAsync();
                if (result == null)
                {
                    await transaction.RollbackAsync();
                    return Results.Conflict("Failed to insert park - no ID returned");
                }
                parkId = Convert.ToInt32(result);
            }

            // Insert activities with parameterized queries
            var insertActivitySql = @"
                INSERT INTO activities (park_id, name, description)
                VALUES (@parkId, @name, @description)";

            foreach (Activity activity in park.Activities)
            {
                await using var cmd = new NpgsqlCommand(insertActivitySql, conn, transaction);
                cmd.Parameters.AddWithValue("parkId", parkId);
                cmd.Parameters.AddWithValue("name", activity.Name);
                cmd.Parameters.AddWithValue("description", activity.Description ?? "");
                await cmd.ExecuteNonQueryAsync();
            }

            await transaction.CommitAsync();

            return Results.Created<Park>($"/parks/{parkCode}", park);
        }
        catch (Exception ex)
        {
            await transaction.RollbackAsync();
            return Results.Problem($"Error creating park: {ex.Message}");
        }
    }

    public async Task<IResult> GetParkAsync(string parkCode)
    {
        Park park;
        await using var conn = new NpgsqlConnection(_connectionString);
        await conn.OpenAsync();

        var sql = @"SELECT
            p.id,
            p.name,
            p.park_code,
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
            WHERE p.park_code = @parkCode
            GROUP BY p.id, p.name, p.park_code, p.state_code, p.latitude, p.longitude;";

        await using var cmd = new NpgsqlCommand(sql, conn);
        cmd.Parameters.AddWithValue("parkCode", parkCode);

        await using var reader = await cmd.ExecuteReaderAsync();

        while (await reader.ReadAsync())
        {
            park = new Park
            {
                Id = reader.GetInt32(0).ToString(),
                Name = reader.GetString(1),
                ParkCode = reader.GetString(2),
                StateCode = reader.GetString(3),
                Latitude = reader.GetFloat(4),
                Longitude = reader.GetFloat(5),
                Activities = System.Text.Json.JsonSerializer.Deserialize<Activity[]>(reader.GetString(6)) ??
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
              p.park_code,
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
          GROUP BY p.id, p.name, p.park_code, p.state_code, p.latitude, p.longitude";

        await using var cmd = new NpgsqlCommand(sql, conn);
        await using var reader = await cmd.ExecuteReaderAsync();

        while (await reader.ReadAsync())
        {
            parks.Add(new Park
            {
                Id = reader.GetInt32(0).ToString(),
                Name = reader.GetString(1),
                ParkCode = reader.GetString(2),
                StateCode = reader.GetString(3),
                Latitude = reader.GetFloat(4),
                Longitude = reader.GetFloat(5),
                Activities = System.Text.Json.JsonSerializer.Deserialize<Activity[]>(reader.GetString(6)) ??
    Array.Empty<Activity>()
            });
        }


        return Results.Ok(parks);
    }

    public async Task<IResult> UpdateParkAsync(string parkCode, Park park)
    {
        await using var conn = new NpgsqlConnection(_connectionString);
        await conn.OpenAsync();

        // Start transaction
        await using var transaction = await conn.BeginTransactionAsync();

        try
        {
            // First, get the park ID by parkCode
            var getParkIdSql = "SELECT id FROM parks WHERE park_code = @parkCode";
            await using var getParkCmd = new NpgsqlCommand(getParkIdSql, conn, transaction);
            getParkCmd.Parameters.AddWithValue("parkCode", parkCode);

            var parkId = await getParkCmd.ExecuteScalarAsync();

            if (parkId == null)
            {
                await transaction.RollbackAsync();
                return Results.NotFound($"Park with code '{parkCode}' not found");
            }

            // Update park fields
            var updateParkSql = @"
                UPDATE parks
                SET name = @name,
                    state_code = @stateCode,
                    latitude = @latitude,
                    longitude = @longitude
                WHERE park_code = @parkCode";

            await using var updateCmd = new NpgsqlCommand(updateParkSql, conn, transaction);
            updateCmd.Parameters.AddWithValue("name", park.Name);
            updateCmd.Parameters.AddWithValue("stateCode", park.StateCode);
            updateCmd.Parameters.AddWithValue("latitude", park.Latitude);
            updateCmd.Parameters.AddWithValue("longitude", park.Longitude);
            updateCmd.Parameters.AddWithValue("parkCode", parkCode);

            await updateCmd.ExecuteNonQueryAsync();

            // Delete existing activities
            var deleteActivitiesSql = "DELETE FROM activities WHERE park_id = @parkId";
            await using var deleteCmd = new NpgsqlCommand(deleteActivitiesSql, conn, transaction);
            deleteCmd.Parameters.AddWithValue("parkId", parkId);
            await deleteCmd.ExecuteNonQueryAsync();

            // Insert new activities
            foreach (Activity activity in park.Activities)
            {
                var insertActivitySql = @"
                    INSERT INTO activities (park_id, name, description)
                    VALUES (@parkId, @name, @description)";

                await using var insertCmd = new NpgsqlCommand(insertActivitySql, conn, transaction);
                insertCmd.Parameters.AddWithValue("parkId", parkId);
                insertCmd.Parameters.AddWithValue("name", activity.Name);
                insertCmd.Parameters.AddWithValue("description", activity.Description ?? "");
                await insertCmd.ExecuteNonQueryAsync();
            }

            // Commit transaction
            await transaction.CommitAsync();

            return Results.Ok(park);
        }
        catch (Exception ex)
        {
            await transaction.RollbackAsync();
            return Results.Problem($"Error updating park: {ex.Message}");
        }
    }

    public async Task<IResult> DeleteParkAsync(string parkCode)
    {
        var sql = "DELETE FROM parks WHERE park_code = @parkCode RETURNING id";

        await using var conn = new NpgsqlConnection(_connectionString);
        await conn.OpenAsync();

        await using var cmd = new NpgsqlCommand(sql, conn);
        cmd.Parameters.AddWithValue("parkCode", parkCode);

        var deletedId = await cmd.ExecuteScalarAsync();

        if (deletedId == null)
        {
            return Results.NotFound($"Park with code '{parkCode}' not found");
        }

        return Results.NoContent();
    }

    public async Task<IResult> SearchGeographic(double latitude, double longitude, string activity, double radiusKm)
    {
        List<Park> parks = new();

        await using var conn = new NpgsqlConnection(_connectionString);
        await conn.OpenAsync();

        var sql = @"SELECT
            p.id,
            p.name,
            p.park_code,
            p.state_code,
            p.latitude,
            p.longitude,
            COALESCE(
                json_agg(
                    json_build_object('Name', a.name, 'description', a.description)
                ) FILTER (WHERE a.id IS NOT NULL),
                '[]'
            ) as activities,
            ST_Distance(location, ST_MakePoint(@longitude, @latitude)::geography) / 1000 AS distance_km
            FROM parks p
            LEFT JOIN activities a ON p.id = a.park_id
            WHERE
            to_tsvector('english', a.name) @@ to_tsquery('english', @activity)
            AND ST_DWithin(location, ST_MakePoint(@longitude, @latitude)::geography, @radiusMeters)
            GROUP BY p.id, p.name, p.park_code, p.state_code, p.latitude, p.longitude, distance_km
            ORDER BY distance_km;";

        await using var cmd = new NpgsqlCommand(sql, conn);
        cmd.Parameters.AddWithValue("latitude", latitude);
        cmd.Parameters.AddWithValue("longitude", longitude);
        cmd.Parameters.AddWithValue("activity", activity);
        cmd.Parameters.AddWithValue("radiusMeters", radiusKm * 1000);

        await using var reader = await cmd.ExecuteReaderAsync();

        while (await reader.ReadAsync())
        {
            parks.Add(new Park
            {
                Id = reader.GetInt32(0).ToString(),
                Name = reader.GetString(1),
                ParkCode = reader.GetString(2),
                StateCode = reader.GetString(3),
                Latitude = reader.GetFloat(4),
                Longitude = reader.GetFloat(5),
                Activities = System.Text.Json.JsonSerializer.Deserialize<Activity[]>(reader.GetString(6)) ??
                    Array.Empty<Activity>()
            });
        }

        return Results.Ok(parks);
    }
}
