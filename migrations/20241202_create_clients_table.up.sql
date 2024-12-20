CREATE TABLE clients (
    u_id VARCHAR(50) PRIMARY KEY,
    client_name VARCHAR(255) NOT NULL,
    client_appkey VARCHAR(255) NOT NULL UNIQUE,
    client_secret VARCHAR(255) NOT NULL UNIQUE,
    client_appid VARCHAR(255) NOT NULL UNIQUE,
    app_name VARCHAR(255) NOT NULL,
    mobile VARCHAR(50) NOT NULL,
    client_status INT NOT NULL,
    testing INT NOT NULL,
    lang VARCHAR(10) NOT NULL,
    callback_url VARCHAR(255) NOT NULL,
    fail_callback VARCHAR(10) NOT NULL,
    isdcb VARCHAR(10) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
