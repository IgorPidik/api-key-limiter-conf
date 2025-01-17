CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    oauth2_id INT NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL
);

