/* phone_data */

CREATE TABLE registration (
    id BIGINT PRIMARY KEY,
    phone_number VARCHAR(50),
    created_at TIMESTAMP
);

CREATE SEQUENCE registration_id_seq;
ALTER TABLE registration ALTER COLUMN id SET DEFAULT nextval('registration_id_seq');

CREATE TABLE transaction (
    id BIGINT PRIMARY KEY,
    phone_number VARCHAR(50),
    transaction_date TIMESTAMP NOT NULL,
    total_deposit BIGINT NOT NULL,
    total_profit BIGINT NOT NULL
);

CREATE SEQUENCE transaction_id_seq;
ALTER TABLE transaction ALTER COLUMN id SET DEFAULT nextval('transaction_id_seq');

CREATE TABLE data (
    id BIGINT PRIMARY KEY,
    whatsapp VARCHAR(50),
    name VARCHAR(50),
    nik VARCHAR(50),
    id_client BIGINT REFERENCES client(id),
    created_at TIMESTAMP
);

CREATE SEQUENCE data_id_seq;
ALTER TABLE data ALTER COLUMN id SET DEFAULT nextval('data_id_seq');

CREATE TABLE client (
    id BIGINT PRIMARY KEY,
    name VARCHAR(50),
    created_at TIMESTAMP
);

CREATE SEQUENCE client_id_seq;
ALTER TABLE client ALTER COLUMN id SET DEFAULT nextval('client_id_seq');

CREATE SEQUENCE IF NOT EXISTS admin_id_seq
START WITH 1
INCREMENT BY 1
NO MINVALUE
NO MAXVALUE
CACHE 1;

-- 2. Tabel admin
CREATE TABLE public.admin (
    id        integer NOT NULL DEFAULT nextval('admin_id_seq'),
    username  character varying(30) NOT NULL,
    password  bytea NOT NULL,
    role      text NOT NULL DEFAULT 'user',
    "isActive" boolean NOT NULL DEFAULT true,
    "createdAt" timestamp(3) without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" timestamp(3) without time zone NOT NULL,

    CONSTRAINT admin_pkey PRIMARY KEY (id)
);

-- Sequence untuk primary key
CREATE SEQUENCE IF NOT EXISTS activity_log_id_seq
START WITH 1
INCREMENT BY 1
NO MINVALUE
NO MAXVALUE
CACHE 1;

-- Tabel activity_log
CREATE TABLE activity_log (
    id BIGINT NOT NULL DEFAULT nextval('activity_log_id_seq'::regclass),
    admin_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    details TEXT,
    created_at TIMESTAMP(3) WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT activity_log_pkey PRIMARY KEY (id),
    CONSTRAINT activity_log_admin_id_fkey
        FOREIGN KEY (admin_id)
        REFERENCES admin(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);


--Opsional, jika ingin membuat unique constraint pada kolom whatsapp
-- ALTER TABLE data
-- ADD CONSTRAINT data_whatsapp_unique UNIQUE (whatsapp);

