-- Add full-text search column for events_archive
ALTER TABLE events_archive 
ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Create function to update search vector
CREATE OR REPLACE FUNCTION update_events_search_vector() 
RETURNS trigger AS $$
BEGIN
    NEW.search_vector := 
        setweight(to_tsvector('russian', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('russian', COALESCE(NEW.description, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update search vector
DROP TRIGGER IF EXISTS update_events_search_vector_trigger ON events_archive;
CREATE TRIGGER update_events_search_vector_trigger
    BEFORE INSERT OR UPDATE ON events_archive
    FOR EACH ROW EXECUTE FUNCTION update_events_search_vector();

-- Update existing records with search vectors
UPDATE events_archive SET search_vector = 
    setweight(to_tsvector('russian', COALESCE(title, '')), 'A') ||
    setweight(to_tsvector('russian', COALESCE(description, '')), 'B');

-- Create GIN index for fast full-text search
CREATE INDEX IF NOT EXISTS events_search_vector_idx 
ON events_archive USING GIN(search_vector);

-- Create additional index for Russian text search
CREATE INDEX IF NOT EXISTS events_title_description_fts_idx 
ON events_archive USING GIN(
    (setweight(to_tsvector('russian', title), 'A') || 
     setweight(to_tsvector('russian', COALESCE(description, '')), 'B'))
);`