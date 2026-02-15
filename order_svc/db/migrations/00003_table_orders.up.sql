CREATE TABLE orders (
    id          UUID         PRIMARY KEY,
    items       JSONB        NOT NULL,
    status      order_status NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
