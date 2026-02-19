-- Завершённые поездки (создаётся room_service при COMPLETED)
CREATE TABLE IF NOT EXISTS public.routes (
    route_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id     UUID        NOT NULL UNIQUE,
    driver_id   UUID        NOT NULL,
    start_point TEXT        NOT NULL,
    end_point   TEXT        NOT NULL,
    distance    NUMERIC(10,2) NOT NULL DEFAULT 0,
    total_price NUMERIC(10,2) NOT NULL DEFAULT 0,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Связь пассажир ↔ поездка
CREATE TABLE IF NOT EXISTS public.room_passengers (
    route_id UUID NOT NULL REFERENCES public.routes(route_id) ON DELETE CASCADE,
    user_id  UUID NOT NULL,
    paid     BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (route_id, user_id)
);

CREATE INDEX IF NOT EXISTS room_passengers_user_idx ON public.room_passengers(user_id);
CREATE INDEX IF NOT EXISTS routes_driver_idx ON public.routes(driver_id);
