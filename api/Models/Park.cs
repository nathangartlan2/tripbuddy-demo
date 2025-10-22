using System.Text.Json.Serialization;
namespace api.Models;

public class Park
{
    [JsonPropertyName("name")]
    public string Name { get; set; }

    [JsonPropertyName("Id")]
    public string Id { get; set; } = "";

    [JsonPropertyName("stateCode")]
    public string StateCode { get; set; }

    [JsonPropertyName("latitude")]
    public float Latitude { get; set; }

    [JsonPropertyName("longitude")]
    public float Longitude { get; set; }
}
