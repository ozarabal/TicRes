CREATE TABLE transactions (
  payment_id SERIAL PRIMARY KEY,
  amount DECIMAL(10, 2),
  payment_method VARCHAR(50),
  booking_id INTEGER UNIQUE, -- UNIQUE karena relasinya One-to-One ( - )
  transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  external_id VARCHAR(255),

  -- Foreign Key
  CONSTRAINT fk_transaction_booking
    FOREIGN KEY (booking_id)
    REFERENCES booking (booking_id)
);