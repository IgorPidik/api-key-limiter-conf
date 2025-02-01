CREATE TYPE LIMIT_DURATION AS ENUM ('second', 'minute', 'hour', 'day', 'week', 'month', 'year', 'forever');

CREATE TABLE configs (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    project_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    limit_requests_count INT NOT NULL,
    limit_duration LIMIT_DURATION NOT NULL,
    CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE
);
