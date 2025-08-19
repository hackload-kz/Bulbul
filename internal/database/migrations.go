package database

import (
	"fmt"
	"log"
)

func (db *DB) RunMigrations() error {
	log.Println("Running database migrations...")

	migrations := []string{
		createUsersTable,
		createEventsTable,
		createSeatsTable,
		createBookingsTable,
		createBookingSeatsTable,
		// addFullTextSearchToEvents,
	}

	for i, migration := range migrations {
		log.Printf("Running migration %d...", i+1)
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	log.Println("All migrations completed successfully")
	return nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(64) NOT NULL,
    password_plain VARCHAR(255),
    first_name VARCHAR(100) NOT NULL,
    surname VARCHAR(100) NOT NULL,
    birthday DATE,
    registered_at TIMESTAMP NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_logged_in TIMESTAMP NOT NULL DEFAULT NOW()
);`

const createEventsTable = `
CREATE TABLE IF NOT EXISTS events_archive (
    id SERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    datetime_start TIMESTAMP NOT NULL,
    provider VARCHAR(100) NOT NULL,
    external BOOLEAN NOT NULL DEFAULT FALSE,
    total_seats INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);`

const createSeatsTable = `
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS seats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id INTEGER NOT NULL REFERENCES events_archive(id) ON DELETE CASCADE,
    row_number INTEGER NOT NULL,
    seat_number INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'FREE',
    price DECIMAL(10,2),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(event_id, row_number, seat_number),
    CHECK (status IN ('FREE', 'RESERVED', 'SOLD'))
);`

const createBookingsTable = `
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL REFERENCES events_archive(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(user_id),
    status VARCHAR(20) NOT NULL DEFAULT 'CREATED',
    payment_status VARCHAR(20) DEFAULT 'PENDING',
    total_amount DECIMAL(10,2) DEFAULT 0,
    payment_id VARCHAR(255),
    order_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CHECK (status IN ('CREATED', 'CONFIRMED', 'CANCELLED', 'EXPIRED')),
    CHECK (payment_status IN ('PENDING', 'INITIATED', 'COMPLETED', 'FAILED', 'CANCELLED'))
);`

const createBookingSeatsTable = `
CREATE TABLE IF NOT EXISTS booking_seats (
    id SERIAL PRIMARY KEY,
    booking_id INTEGER NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    seat_id UUID NOT NULL REFERENCES seats(id) ON DELETE CASCADE,
    reserved_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(booking_id, seat_id)
);`

const addFullTextSearchToEvents = `
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
