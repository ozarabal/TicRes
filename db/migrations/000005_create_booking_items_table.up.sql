CREATE TABLE booking_items (
  id SERIAL PRIMARY KEY,
  booking_id INTEGER,
  seat_id INTEGER,

  -- Foreign Keys
  CONSTRAINT fk_booking_items_booking
    FOREIGN KEY (booking_id)
    REFERENCES booking (booking_id),
    
  CONSTRAINT fk_booking_items_seats
    FOREIGN KEY (seat_id)
    REFERENCES seats (seat_id)
);