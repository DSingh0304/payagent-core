CREATE TABLE products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    category    VARCHAR(100),
    price_paise BIGINT NOT NULL,
    stock       INTEGER NOT NULL DEFAULT 0,
    tags        TEXT[],
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE orders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id          VARCHAR(255) NOT NULL,
    razorpay_order_id   VARCHAR(255) UNIQUE,
    status              VARCHAR(50) NOT NULL DEFAULT 'pending',
    amount_paise        BIGINT NOT NULL,
    items               JSONB NOT NULL,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    session_id  VARCHAR(255) NOT NULL,
    event_type  VARCHAR(100) NOT NULL,
    actor       VARCHAR(50) NOT NULL, 
    payload     JSONB,
    reasoning   TEXT,
    outcome     VARCHAR(50),
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE upi_mandates (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id          VARCHAR(255) NOT NULL,
    customer_id         VARCHAR(255), 
    token_id            VARCHAR(255),
    block_order_id      VARCHAR(255),
    max_amount_paise    BIGINT DEFAULT 1000000,
    status              VARCHAR(50) DEFAULT 'pending',
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO products (name, description, category, price_paise, stock, tags) VALUES
('Nike Air Max 270', 'Lightweight running shoe', 'shoes', 289900, 10, ARRAY['running','nike']),
('Campus Hype', 'Budget casual sneaker', 'shoes', 149900, 20, ARRAY['casual','budget']),
('boAt Airdopes 141', 'True wireless earbuds', 'electronics', 149900, 25, ARRAY['earbuds']);
