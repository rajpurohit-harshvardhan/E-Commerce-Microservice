CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) on delete cascade,
    product_id UUID NOT NULL,
    quantity INT8 NOT NULL,
    price DECIMAL(10,2) NOT NULL
);