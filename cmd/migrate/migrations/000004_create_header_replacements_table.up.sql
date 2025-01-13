CREATE TABLE header_replacements (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    config_id UUID NOT NULL,
    header_name VARCHAR(255) NOT NULL,
    header_value VARCHAR(255) NOT NULL,
    CONSTRAINT fk_config FOREIGN KEY (config_id) REFERENCES configs (id) ON DELETE CASCADE
);
