-- Enable PostGIS extension for geographic data
CREATE EXTENSION IF NOT EXISTS postgis;

-- Enable pg_trgm extension for autocomplete/fuzzy search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create parks table
CREATE TABLE parks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    park_code VARCHAR(255) UNIQUE NOT NULL,
    park_url VARCHAR(500),
    state_code VARCHAR(2) NOT NULL,
    latitude DECIMAL(10, 6) NOT NULL,
    longitude DECIMAL(11, 6) NOT NULL,
    -- PostGIS geography column (automatically generated from lat/lon)
    location GEOGRAPHY(POINT, 4326),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create activities table (normalized design for better querying)
CREATE TABLE activities (
    id SERIAL PRIMARY KEY,
    park_id INTEGER NOT NULL REFERENCES parks(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for optimal search performance

-- Spatial index for geographic queries (find parks near a location)
CREATE INDEX idx_parks_location ON parks USING GIST(location);

-- Index for filtering by state
CREATE INDEX idx_parks_state ON parks(state_code);

-- Index for park_code lookups (unique constraint already creates an index, but explicit for clarity)
-- Note: UNIQUE constraint automatically creates an index, so this is optional
-- CREATE INDEX idx_parks_code ON parks(park_code);

-- Full-text search index on park names
CREATE INDEX idx_parks_name_fts ON parks USING GIN(to_tsvector('english', name));

-- Trigram index for autocomplete and fuzzy matching
CREATE INDEX idx_parks_name_trgm ON parks USING GIN(name gin_trgm_ops);

-- Index for activity lookups
CREATE INDEX idx_activities_park_id ON activities(park_id);
CREATE INDEX idx_activities_name_trgm ON activities USING GIN(name gin_trgm_ops);

-- Function to automatically update the location column when lat/lon changes
CREATE OR REPLACE FUNCTION update_park_location()
RETURNS TRIGGER AS $$
BEGIN
    NEW.location = ST_SetSRID(ST_MakePoint(NEW.longitude, NEW.latitude), 4326)::geography;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update location on insert or update
CREATE TRIGGER trigger_update_park_location
BEFORE INSERT OR UPDATE ON parks
FOR EACH ROW
EXECUTE FUNCTION update_park_location();

-- Function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update updated_at timestamp
CREATE TRIGGER trigger_update_parks_updated_at
BEFORE UPDATE ON parks
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Create a view for parks with activity counts (useful for queries)
CREATE VIEW parks_with_activity_count AS
SELECT
    p.id,
    p.name,
    p.park_code,
    p.park_url,
    p.state_code,
    p.latitude,
    p.longitude,
    p.location,
    COUNT(a.id) as activity_count,
    p.created_at,
    p.updated_at
FROM parks p
LEFT JOIN activities a ON p.id = a.park_id
GROUP BY p.id, p.name, p.park_code, p.park_url, p.state_code, p.latitude, p.longitude, p.location, p.created_at, p.updated_at;

-- Grant permissions (if needed for specific user)
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO tripbuddy_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO tripbuddy_user;
