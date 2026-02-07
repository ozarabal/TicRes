CREATE TABLE refund(
    refund_id SERIAL PRIMARY KEY,
    booking_id INTEGER,
    amount DECIMAL(10,2),
    refund_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reason VARCHAR(255),

    CONSTRAINT fk_refund_booking
    FOREIGN KEY (booking_id)
    REFERENCES booking (booking_id)
);