CREATE TABLE IF NOT EXISTS rooms (
    room_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id      UUID NOT NULL,
    start_latitude  DOUBLE PRECISION NOT NULL,
    start_longitude DOUBLE PRECISION NOT NULL,
    end_latitude    DOUBLE PRECISION NOT NULL,
    end_longitude   DOUBLE PRECISION NOT NULL,
    available_seats INT NOT NULL,
    status          INT NOT NULL DEFAULT 1,
    total_price     REAL NOT NULL DEFAULT 0,
    cost_per_member REAL NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    scheduled_time  TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS room_members (
    room_id UUID NOT NULL REFERENCES rooms(room_id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (room_id, user_id)
);

CREATE INDEX IF NOT EXISTS room_members_user_idx ON room_members(user_id);
CREATE INDEX IF NOT EXISTS rooms_status_idx ON rooms(status);
