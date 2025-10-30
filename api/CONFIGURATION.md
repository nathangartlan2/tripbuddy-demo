# TripBuddy API Configuration Guide

This document explains how to configure secrets and connection strings for different environments.

## Local Development (dotnet run)

### Option 1: User Secrets (Recommended for Sensitive Data)

User Secrets keep sensitive data out of your code and appsettings files.

```bash
# Initialize user secrets (one-time setup)
dotnet user-secrets init

# Set the PostgreSQL connection string
dotnet user-secrets set "ConnectionStrings:PostgreSQL" "Host=localhost;Port=5432;Database=tripbuddy;Username=tripbuddy_user;Password=YOUR_PASSWORD_HERE"

# View all secrets
dotnet user-secrets list

# Remove a secret
dotnet user-secrets remove "ConnectionStrings:PostgreSQL"
```

**Advantages:**
- ✅ Never committed to source control
- ✅ Stored outside project directory
- ✅ Perfect for real passwords/API keys

### Option 2: appsettings.Development.json (Current Setup)

The connection string is currently in `appsettings.Development.json` for convenience.

**⚠️ Warning:** This file is committed to git! Only use for:
- Local development databases with non-sensitive passwords
- Shared development environments

**For production-like security in development:**
1. Remove `ConnectionStrings` section from `appsettings.Development.json`
2. Add to `.gitignore`: `appsettings.Development.json`
3. Use User Secrets instead (Option 1 above)

---

## Docker Development (docker-compose)

Connection string is set via environment variable in `docker-compose.yml`:

```yaml
environment:
  - ConnectionStrings__PostgreSQL=Host=postgres;Port=5432;Database=tripbuddy;Username=tripbuddy_user;Password=tripbuddy_pass
```

**For better security:**
1. Create a `.env` file (add to `.gitignore`):
   ```env
   POSTGRES_PASSWORD=tripbuddy_pass
   ```

2. Update `docker-compose.yml` to reference it:
   ```yaml
   environment:
     - ConnectionStrings__PostgreSQL=Host=postgres;Port=5432;Database=tripbuddy;Username=tripbuddy_user;Password=${POSTGRES_PASSWORD}
   ```

---

## Production (Fly.io)

Fly.io uses secrets for sensitive configuration.

### Setting Up PostgreSQL on Fly.io

```bash
# 1. Create a Postgres database
fly postgres create --name tripbuddy-db

# 2. Attach it to your app (automatically sets connection string)
fly postgres attach tripbuddy-db -a tripbuddy-api
```

This automatically creates a secret called `DATABASE_URL`. To use it:

```bash
# View current secrets
fly secrets list

# The connection string is automatically set when you attach the database
# But if you need to set it manually:
fly secrets set ConnectionStrings__PostgreSQL="Host=tripbuddy-db.internal;Port=5432;Database=tripbuddy;Username=postgres;Password=GENERATED_PASSWORD"
```

### Other Secrets for Production

```bash
# Set additional secrets (if needed)
fly secrets set MAPBOX_API_KEY=your_mapbox_key_here

# View all secrets (values are hidden)
fly secrets list

# Remove a secret
fly secrets unset SOME_SECRET_NAME
```

---

## Configuration Priority (ASP.NET Core)

ASP.NET Core loads configuration in this order (later sources override earlier):

1. `appsettings.json` (base configuration)
2. `appsettings.{Environment}.json` (environment-specific)
3. User Secrets (Development environment only)
4. Environment Variables
5. Command-line arguments

**Example:**
- If `ConnectionStrings:PostgreSQL` is in both `appsettings.Development.json` and User Secrets, User Secrets wins
- Environment variables always override appsettings files

---

## Environment Variable Format

When using environment variables, use double underscores `__` to represent nested JSON:

**JSON:**
```json
{
  "ConnectionStrings": {
    "PostgreSQL": "Host=..."
  }
}
```

**Environment Variable:**
```bash
ConnectionStrings__PostgreSQL=Host=...
```

---

## Security Best Practices

### ✅ DO:
- Use User Secrets for local development with real passwords
- Use Fly.io Secrets for production
- Use environment variables in Docker
- Add `.env` files to `.gitignore`
- Rotate passwords regularly
- Use strong passwords

### ❌ DON'T:
- Commit passwords to git (even in private repos)
- Use the same password across environments
- Share production credentials in Slack/Email
- Hardcode secrets in source code

---

## Checking Configuration

### At Runtime:
```csharp
// In Program.cs - this will throw an exception if not configured
string connection = builder.Configuration.GetConnectionString("PostgreSQL")
    ?? throw new InvalidOperationException("Connection string not configured");
```

### From Command Line:
```bash
# Check what configuration is loaded (sanitized output)
dotnet run --urls="http://localhost:5000"

# If connection string is missing, you'll see:
# System.InvalidOperationException: PostgreSQL connection string is not configured
```

---

## Troubleshooting

### "PostgreSQL connection string is not configured"

**Cause:** No connection string found in any configuration source.

**Solutions:**
1. Check `appsettings.Development.json` has `ConnectionStrings:PostgreSQL`
2. Or set via User Secrets: `dotnet user-secrets set "ConnectionStrings:PostgreSQL" "..."`
3. Or set environment variable: `export ConnectionStrings__PostgreSQL="..."`

### Docker container can't connect to database

**Cause:** Connection string points to `localhost` instead of container name.

**Solution:** In docker-compose, use `Host=postgres` (the service name), not `Host=localhost`

### Fly.io deployment fails with database connection error

**Cause:** Database not attached or wrong connection string.

**Solution:**
```bash
# Check if database is attached
fly postgres attach tripbuddy-db -a tripbuddy-api

# Verify secrets
fly secrets list
```

---

## Quick Reference

| Environment | Configuration Method | File/Command |
|------------|---------------------|--------------|
| Local Dev | appsettings | `appsettings.Development.json` |
| Local Dev (Secure) | User Secrets | `dotnet user-secrets set ...` |
| Docker | Environment Variables | `docker-compose.yml` |
| Fly.io | Fly Secrets | `fly secrets set ...` |

---

*Last Updated: 2025-10-30*
