CREATE TABLE
    safe_regions (
        id UUID PRIMARY KEY,
        user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        group_id UUID NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
        name TEXT NOT NULL,
        latitude DOUBLE PRECISION NOT NULL CHECK (latitude BETWEEN -90 AND 90),
        longitude DOUBLE PRECISION NOT NULL CHECK (longitude BETWEEN -180 AND 180),
        radius_meters INTEGER NOT NULL CHECK (radius_meters > 0),
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE INDEX idx_safe_regions_user_id ON safe_regions (user_id);

CREATE INDEX idx_safe_regions_group_id ON safe_regions (group_id);