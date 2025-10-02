-- +goose Up
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL NOT NULL
    );

-- +goose Down
DROP TABLE IF EXISTS order_items;
