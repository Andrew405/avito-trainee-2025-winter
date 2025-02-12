CREATE TABLE IF NOT EXISTS inventory (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    item_type VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_user_inventory
    FOREIGN KEY(user_id)
    REFERENCES users(id)
    ON DELETE CASCADE,
    CONSTRAINT unique_item_per_user UNIQUE (user_id, item_type)
);