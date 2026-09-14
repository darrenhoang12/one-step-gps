CREATE TABLE device_preferences (
    device_id TEXT PRIMARY KEY,
    sort_order INTEGER,
    archived BOOLEAN NOT NULL DEFAULT FALSE,
    hidden BOOLEAN NOT NULL DEFAULT FALSE,
    custom_display_name TEXT,
    icon_storage_path TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);