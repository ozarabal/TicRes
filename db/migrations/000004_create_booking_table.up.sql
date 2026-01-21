CREATE TABLE booking (
  booking_id SERIAL PRIMARY KEY,
  user_id INTEGER,
  event_id INTEGER,
  status VARCHAR(50),
  total_amount DECIMAL(10, 2), -- Menggunakan DECIMAL untuk mata uang
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP,

  -- Foreign Keys
  CONSTRAINT fk_bookings_users
    FOREIGN KEY (user_id)
    REFERENCES users (user_id),
  
  CONSTRAINT fk_bookings_event
    FOREIGN KEY (event_id)
    REFERENCES events (event_id)
);