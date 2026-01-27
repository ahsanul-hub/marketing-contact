/* phone_data */

CREATE TABLE client (
    id BIGINT PRIMARY KEY DEFAULT nextval('client_id_seq'),
    name VARCHAR(50),
    created_at TIMESTAMP
);

CREATE TABLE registration (
    id BIGINT PRIMARY KEY DEFAULT nextval('registration_id_seq'),
    phone_number VARCHAR(50),
    id_client BIGINT NOT NULL,
    created_at TIMESTAMP,

    CONSTRAINT registration_id_client_fkey
      FOREIGN KEY (id_client)
      REFERENCES client(id)
      ON UPDATE CASCADE
      ON DELETE RESTRICT
);

CREATE TABLE transaction (
    id BIGINT PRIMARY KEY DEFAULT nextval('transaction_id_seq'),
    phone_number VARCHAR(50),
    id_client BIGINT NOT NULL,
    transaction_date TIMESTAMP NOT NULL,
    total_deposit BIGINT NOT NULL,
    total_profit BIGINT NOT NULL,

    CONSTRAINT transaction_id_client_fkey
      FOREIGN KEY (id_client)
      REFERENCES client(id)
      ON UPDATE CASCADE
      ON DELETE RESTRICT
);

CREATE TABLE data (
    id BIGINT PRIMARY KEY DEFAULT nextval('data_id_seq'),
    whatsapp VARCHAR(50),
    name VARCHAR(50),
    nik VARCHAR(50),
    id_client BIGINT,
    created_at TIMESTAMP,

    CONSTRAINT data_id_client_fkey
      FOREIGN KEY (id_client)
      REFERENCES client(id)
      ON UPDATE CASCADE
      ON DELETE SET NULL
);

-- 2. Tabel admin untuk login
CREATE TABLE "user" (
    id SERIAL PRIMARY KEY,
    username VARCHAR(30) NOT NULL UNIQUE,
    password BYTEA NOT NULL,
    role TEXT NOT NULL DEFAULT 'client',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL
);


CREATE TABLE activity_log (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    details TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT activity_log_user_id_fkey
      FOREIGN KEY (user_id)
      REFERENCES "user"(id)
      ON UPDATE CASCADE
      ON DELETE RESTRICT
);



--Opsional, jika ingin membuat unique constraint pada kolom whatsapp
-- ALTER TABLE data
-- ADD CONSTRAINT data_whatsapp_unique UNIQUE (whatsapp);
