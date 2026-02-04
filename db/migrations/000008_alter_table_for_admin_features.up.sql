CREATE TYPE user_role AS ENUM ('admin', 'user');

CREATE TYPE status_event as ENUM ('available','cancelled', 'completed')

ALTER TABLE users ADD COLUMN role user_role DEFAULT 'user';

ALTER TABLE event ADD COLUMN status status_event, ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

