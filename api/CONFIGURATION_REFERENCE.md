# Configuration Reference

## Current Configuration Values

### Required Configuration

| Setting | Type | Default | Description | Secret? |
|---------|------|---------|-------------|---------|
| `RepositoryType` | string | `PostgreSQL` | Data store to use: `File` or `PostgreSQL` | No |
| `ConnectionStrings:PostgreSQL` | string | - | PostgreSQL connection string | **YES** |
| `DATABASE_URL` | string | - | Fly.io managed Postgres URL (production) | **YES** |

### Optional Configuration

| Setting | Type | Default | Description | Secret? |
|---------|------|---------|-------------|---------|
| `ASPNETCORE_ENVIRONMENT` | string | `Production` | Environment: `Development`, `Docker`, `Production` | No |
| `ASPNETCORE_URLS` | string | from appsettings | URL bindings for Kestrel | No |

---

## Configuration by Environment

### Development (`dotnet run`)
```json
// appsettings.Development.json
{
  "RepositoryType": "PostgreSQL",
  "ConnectionStrings": {
    "PostgreSQL": "Host=localhost;Port=5432;Database=tripbuddy;Username=tripbuddy_user;Password=tripbuddy_pass"
  }
}
```

**Better (User Secrets):**
```bash
dotnet user-secrets set "ConnectionStrings:PostgreSQL" "Host=localhost;..."
```

### Docker (`docker-compose up`)
```yaml
# docker-compose.yml
environment:
  - ASPNETCORE_ENVIRONMENT=Development
  - ConnectionStrings__PostgreSQL=Host=postgres;Port=5432;Database=tripbuddy;Username=tripbuddy_user;Password=tripbuddy_pass
```

### Production (Fly.io)
```bash
# Automatically set when you attach Postgres:
fly postgres attach tripbuddy-db

# Or manually:
fly secrets set DATABASE_URL="postgres://..."
```

---

## Future Configuration (When Needed)

### When You Add Authentication
```bash
# Development (User Secrets)
dotnet user-secrets set "BasicAuth:Username" "admin"
dotnet user-secrets set "BasicAuth:Password" "hashed_password_here"

# Production (Fly.io)
fly secrets set BasicAuth__Username="admin"
fly secrets set BasicAuth__Password="hashed_password_here"
```

### When You Add External APIs (MapBox, etc.)
```bash
# Development (User Secrets)
dotnet user-secrets set "MapBox:ApiKey" "your_key_here"

# Production (Fly.io)
fly secrets set MapBox__ApiKey="your_key_here"
```

### When You Add Rate Limiting
```json
// appsettings.json (NOT secret)
{
  "RateLimiting": {
    "PermitLimit": 100,
    "Window": "00:01:00"
  }
}
```

### When You Add CORS Restrictions
```json
// appsettings.json (NOT secret)
{
  "Cors": {
    "AllowedOrigins": [
      "https://yourdomain.com",
      "https://www.yourdomain.com"
    ]
  }
}
```

---

## Secret Management Rules

### ✅ Store as Secrets (User Secrets / Fly.io Secrets):
- Database passwords
- API keys
- Authentication credentials
- JWT secrets
- Any value that grants access

### ❌ OK in appsettings.json (Public):
- Feature flags
- URL configurations
- Rate limits
- Logging levels
- Non-sensitive defaults

---

## Checking Current Configuration

### View all configuration sources
```bash
dotnet run --environment Development
# App will show error messages if required config is missing
```

### View Fly.io secrets
```bash
fly secrets list
```

### View User Secrets
```bash
dotnet user-secrets list
```

---

*Last Updated: 2025-10-30*
