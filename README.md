# tripbuddy-demo

Demo for Go Padawan Talk

## Docker command guide

### Rebuild and Restart API Only

```bash
# Rebuild the API container (without cache for clean build)
docker-compose build --no-cache api

# Restart the API service
docker-compose up -d api
```

Or in one command:
```bash
docker-compose up -d --build api
```

### Other Useful Commands

```bash
# Stop the API only
docker-compose stop api

# Start the API only (without rebuilding)
docker-compose start api

# Restart the API (without rebuilding)
docker-compose restart api

# View API logs
docker-compose logs -f api

# Rebuild and restart with logs visible
docker-compose up --build api
```

### Quick Rebuild After Code Changes

After modifying your C# code in `api/Program.cs` or other files:

```bash
docker-compose up -d --build api
```

This will:
1. Rebuild the API Docker image
2. Recreate the container
3. Start it in detached mode
4. Leave the postgres service running unchanged

The API will be available at `http://localhost:8080`
