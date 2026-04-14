-- ============================================================
-- 000001_init_schema.up.sql
-- Creates the full Warung-Ku schema.
-- Equivalent to running `prisma db push` on the original app.
-- ============================================================

-- users
CREATE TABLE IF NOT EXISTS users (
    id          TEXT        PRIMARY KEY,
    email       TEXT        UNIQUE,
    password    TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    phone       TEXT        NOT NULL UNIQUE,
    birthday    TIMESTAMPTZ,
    photo_url   TEXT,
    role        TEXT        NOT NULL DEFAULT 'MEMBER',
    points      INTEGER     NOT NULL DEFAULT 0,
    address     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- products
CREATE TABLE IF NOT EXISTS products (
    id             TEXT        PRIMARY KEY,
    name           TEXT        NOT NULL,
    description    TEXT,
    sku            TEXT        NOT NULL UNIQUE,
    type           TEXT        NOT NULL DEFAULT 'READY_STOCK',
    is_active      BOOLEAN     NOT NULL DEFAULT TRUE,
    created_by_id  TEXT        NOT NULL REFERENCES users(id),
    updated_by_id  TEXT        REFERENCES users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- product_variants
CREATE TABLE IF NOT EXISTS product_variants (
    id          TEXT            PRIMARY KEY,
    product_id  TEXT            NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name        TEXT            NOT NULL,
    sku         TEXT            NOT NULL UNIQUE,
    price       DECIMAL(10, 2)  NOT NULL DEFAULT 0,
    cost        DECIMAL(10, 2)  NOT NULL DEFAULT 0,
    stock       INTEGER         NOT NULL DEFAULT 0,
    low_stock   INTEGER         NOT NULL DEFAULT 10,
    is_active   BOOLEAN         NOT NULL DEFAULT TRUE,
    points      INTEGER         NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- stock_movements
CREATE TABLE IF NOT EXISTS stock_movements (
    id          TEXT        PRIMARY KEY,
    variant_id  TEXT        NOT NULL REFERENCES product_variants(id),
    quantity    INTEGER     NOT NULL,
    type        TEXT        NOT NULL,
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- customers  (non-member walk-in customers)
CREATE TABLE IF NOT EXISTS customers (
    id          TEXT        PRIMARY KEY,
    name        TEXT        NOT NULL,
    phone       TEXT        NOT NULL UNIQUE,
    address     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- sales
CREATE TABLE IF NOT EXISTS sales (
    id                      TEXT            PRIMARY KEY,
    sale_number             TEXT            NOT NULL UNIQUE,
    customer_id             TEXT            REFERENCES users(id),
    non_member_customer_id  TEXT            REFERENCES customers(id),
    cashier_id              TEXT            NOT NULL REFERENCES users(id),
    subtotal                DECIMAL(10, 2)  NOT NULL DEFAULT 0,
    discount                DECIMAL(10, 2)  NOT NULL DEFAULT 0,
    tax                     DECIMAL(10, 2)  NOT NULL DEFAULT 0,
    ongkir                  DECIMAL(10, 2)  NOT NULL DEFAULT 0,
    total                   DECIMAL(10, 2)  NOT NULL DEFAULT 0,
    payment_method          TEXT            NOT NULL,
    payment_status          TEXT            NOT NULL DEFAULT 'PAID',
    notes                   TEXT,
    points_earned           INTEGER         NOT NULL DEFAULT 0,
    points_redeemed         INTEGER         NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- sale_items
CREATE TABLE IF NOT EXISTS sale_items (
    id          TEXT            PRIMARY KEY,
    sale_id     TEXT            NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    variant_id  TEXT            NOT NULL REFERENCES product_variants(id),
    quantity    INTEGER         NOT NULL,
    price       DECIMAL(10, 2)  NOT NULL DEFAULT 0,
    subtotal    DECIMAL(10, 2)  NOT NULL DEFAULT 0
);

-- point_history
CREATE TABLE IF NOT EXISTS point_history (
    id           TEXT        PRIMARY KEY,
    user_id      TEXT        NOT NULL REFERENCES users(id),
    points       INTEGER     NOT NULL,
    type         TEXT        NOT NULL,
    description  TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ
);

-- cashflows
CREATE TABLE IF NOT EXISTS cashflows (
    id             TEXT            PRIMARY KEY,
    type           TEXT            NOT NULL,
    category       TEXT            NOT NULL,
    amount         DECIMAL(10, 2)  NOT NULL DEFAULT 0,
    description    TEXT            NOT NULL,
    date           TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_by_id  TEXT            NOT NULL REFERENCES users(id),
    created_at     TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- settings
CREATE TABLE IF NOT EXISTS settings (
    id           TEXT        PRIMARY KEY,
    key          TEXT        NOT NULL UNIQUE,
    value        TEXT        NOT NULL,
    description  TEXT,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
