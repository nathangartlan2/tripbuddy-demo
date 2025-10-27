using System.Text.Json.Serialization;
namespace api.Models;


public class Activity
{
    [JsonPropertyName("Name")]
    public string Name { get; set; }

    [JsonPropertyName("description")]
    public string Description { get; set; }
}

public class Park
{

    [JsonPropertyName("id")]
    public string Id { get; set; } = "";

    [JsonPropertyName("name")]
    public string Name { get; set; }

    [JsonPropertyName("park_code")]
    public string ParkCode { get; set; }

    [JsonPropertyName("park_url")]
    public string ParkURL { get; set; } = "";

    [JsonPropertyName("stateCode")]
    public string StateCode { get; set; }

    [JsonPropertyName("latitude")]
    public float Latitude { get; set; }

    [JsonPropertyName("longitude")]
    public float Longitude { get; set; }

    [JsonPropertyName("activities")]
    public Activity[] Activities { get; set; }
}
