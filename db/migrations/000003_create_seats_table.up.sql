CREATE TABLE seats (
  seat_id SERIAL PRIMARY KEY,
  event_id INTEGER,
  seat_number VARCHAR(50),
  category VARCHAR(50),
  is_booked BOOLEAN DEFAULT FALSE,
  version INTEGER DEFAULT 1,
  
  -- Foreign Key ke Events
  CONSTRAINT fk_seats_events
    FOREIGN KEY (event_id)
    REFERENCES events (event_id)
);