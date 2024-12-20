CREATE TABLE payment_method_clients (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    route JSONB,
    status INT,
    msisdn INT,
    client_id VARCHAR(50) NOT NULL
);
