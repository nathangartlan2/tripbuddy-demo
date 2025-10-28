# TripBuddy Database

PostgreSQL database with PostGIS for geographic search and full-text search capabilities.

## Quick Start

```bash
# Start the database
docker-compose up -d postgres

# Check if it's running
docker-compose ps

# View logs
docker-compose logs -f postgres
```

## Database Details

- **Database Name**: `tripbuddy`
- **User**: `tripbuddy_user`
- **Password**: `tripbuddy_pass`
- **Port**: `5432`
- **Extensions**: PostGIS, pg_trgm

## Connection String

For .NET:
```
Host=localhost;Port=5432;Database=tripbuddy;Username=tripbuddy_user;Password=tripbuddy_pass
```

## Schema

### Tables

**parks**
- `id` - Serial primary key
- `name` - Park name (indexed for full-text and trigram search)
- `state_code` - Two-letter state code
- `latitude` / `longitude` - Coordinates
- `location` - PostGIS geography point (auto-generated from lat/lon)
- `created_at` / `updated_at` - Timestamps

**activities**
- `id` - Serial primary key
- `park_id` - Foreign key to parks
- `name` - Activity name
- `description` - Activity description

### Indexes

- Spatial index (GIST) on `location` for geographic queries
- Full-text search (GIN) on park names
- Trigram index (GIN) for autocomplete
- Standard indexes on foreign keys and common filters

## Sample Queries

### Full-Text Search
```sql
-- Find parks with "lake" in the name
SELECT name, state_code
FROM parks
WHERE to_tsvector('english', name) @@ to_tsquery('english', 'lake')
ORDER BY ts_rank(to_tsvector('english', name), to_tsquery('english', 'lake')) DESC;
```

### Geographic Search
```sql
-- Find parks within 50km of Chicago (41.8781, -87.6298)
SELECT
  name,
  ST_Distance(location, ST_MakePoint(-87.6298, 41.8781)::geography) / 1000 AS distance_km
FROM parks
WHERE ST_DWithin(location, ST_MakePoint(-87.6298, 41.8781)::geography, 50000)
ORDER BY distance_km;
```

### Autocomplete
```sql
-- Autocomplete for "star"
SELECT name, similarity(name, 'star') as score
FROM parks
WHERE name % 'star'
ORDER BY score DESC
LIMIT 10;
```

### Combined Search
```sql
-- Find "state park" within 100km of a location
SELECT
  name,
  state_code,
  ST_Distance(location, ST_MakePoint(-87.6298, 41.8781)::geography) / 1000 AS distance_km
FROM parks
WHERE
  to_tsvector('english', name) @@ to_tsquery('english', 'state & park')
  AND ST_DWithin(location, ST_MakePoint(-87.6298, 41.8781)::geography, 100000)
ORDER BY distance_km;
```

## Updating Seed Data

If you update `go-scraper/parks.json`, regenerate the seed data:

```bash
python3 database/generate_seed.py
```

Then restart the database to apply changes:

```bash
docker-compose down
docker volume rm tripbuddy-demo_postgres_data  # WARNING: Deletes all data!
docker-compose up -d postgres
```

## Accessing the Database

### Using psql in Docker
```bash
docker-compose exec postgres psql -U tripbuddy_user -d tripbuddy
```

### Using psql locally (if installed)
```bash
psql -h localhost -p 5432 -U tripbuddy_user -d tripbuddy
```

### Common Commands
```sql
-- List tables
\dt

-- Describe parks table
\d parks

-- Count parks
SELECT COUNT(*) FROM parks;

-- Count activities
SELECT COUNT(*) FROM activities;

-- View parks with activity counts
SELECT * FROM parks_with_activity_count LIMIT 10;
```

## Stopping the Database

```bash
# Stop but keep data
docker-compose stop postgres

# Stop and remove container (keeps data volume)
docker-compose down

# Stop and remove everything including data
docker-compose down -v
```
