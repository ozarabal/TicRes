-- Add status tracking to transactions table
ALTER TABLE transactions ADD COLUMN status VARCHAR(20) DEFAULT 'PENDING';

-- Add status tracking to refund table
ALTER TABLE refund ADD COLUMN status VARCHAR(20) DEFAULT 'PENDING';

-- Add price to seats for calculating total_amount
ALTER TABLE seats ADD COLUMN price DECIMAL(10, 2) DEFAULT 0;
