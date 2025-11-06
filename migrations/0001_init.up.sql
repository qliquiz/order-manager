CREATE TABLE IF NOT EXISTS orders
(
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item     TEXT UNIQUE NOT NULL,
    quantity INTEGER          DEFAULT 0
);
