using System;
using System.Collections;
using api.Models;
using Microsoft.AspNetCore.Mvc;

namespace api.Repositories;

public interface IParksRepository
{
    Task<IResult> GetParksAsync();

    Task<IResult> GetParkAsync(string id);

    Task<IResult> CreateParkAsync(Park park);

    Task<IResult> SearchGeographic(double latitude, double longitude, string activity, double radiusKm);
}
