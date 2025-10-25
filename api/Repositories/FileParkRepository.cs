using System;
using System.Text.Json;
using api.Models;
using Microsoft.AspNetCore.Mvc;

namespace api.Repositories;

public class FileParkRepository : IParksRepository
{
    public static Dictionary<string, Park> Parks { get; private set; } = new();

    // Static constructor - loads parks.json once when class is first used

    private string makeParkKey(string name, string stateCode)
    {
        string cleanName = name.ToLower().Replace(" ", "-").Replace("'", "");
        string cleanState = stateCode.ToLower();
        return $"{cleanName}-{cleanState}";
    }
    public FileParkRepository()
    {
        try
        {
            string jsonPath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "Data", "parks.json");
            Console.WriteLine(jsonPath);
            string jsonContent = File.ReadAllText(jsonPath);
            var parkList = JsonSerializer.Deserialize<List<Park>>(jsonContent);

            if (parkList != null)
            {
                // Build dictionary with composite key: "ParkName_StateCode"
                parkList = parkList.Select(x =>
                {
                    x.Id = makeParkKey(x.Name, x.StateCode);
                    return x;
                }).ToList();

                Parks = parkList.ToDictionary(
                    p => p.Id,
                    p => p
                );
                Console.WriteLine($"[FileParkRepository] Loaded {Parks.Count} parks from parks.json");
            }
        }
        catch (FileNotFoundException)
        {
            Console.WriteLine("[FileParkRepository] parks.json not found, starting with empty repository");
        }
        catch (JsonException ex)
        {
            Console.WriteLine($"[FileParkRepository] Error parsing parks.json: {ex.Message}");
        }
        catch (Exception ex)
        {
            Console.WriteLine($"[FileParkRepository] Error loading parks: {ex.Message}");
        }
    }

    public Task<IResult> CreateParkAsync(Park park)
    {
        throw new NotImplementedException();
    }

    public Task<IResult> GetParkAsync(string id)
    {
        // Lookup park by composite key
        if (Parks.TryGetValue(id, out var park))
        {
            return Task.FromResult(Results.Ok(park));
        }
        return Task.FromResult(Results.NotFound(new { message = $"Park with id '{id}' not found" }));
    }

    public Task<IResult> GetParksAsync()
    {
        // Return all parks
        return Task.FromResult(Results.Ok(Parks.Values));
    }

    public Task<IResult> SearchGeographic(double latitude, double longitude, string activity)
    {
        throw new NotImplementedException();
    }

    public Task<IResult> SearchGeographic(double latitude, double longitude, string activity, double radiusKm)
    {
        throw new NotImplementedException();
    }
}
