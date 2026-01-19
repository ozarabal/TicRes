CREATE TABLE users(
    user_id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    username VARCHAR(30),
    email VARCHAR(255) UNIQUE,
    password VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);