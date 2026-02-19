CREATE TABLE IF NOT EXISTS payments (
    payment_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id             UUID        NOT NULL,
    user_id             UUID        NOT NULL,
    amount              NUMERIC(10,2) NOT NULL,
    currency            VARCHAR(3)  NOT NULL DEFAULT 'RUB',
    status              VARCHAR(30) NOT NULL DEFAULT 'pending',
    yookassa_payment_id VARCHAR(255),
    description         TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS payments_room_idx ON payments(room_id);
CREATE INDEX IF NOT EXISTS payments_user_idx ON payments(user_id);
CREATE INDEX IF NOT EXISTS payments_status_idx ON payments(status);
