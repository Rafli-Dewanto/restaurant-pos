CREATE TABLE orders (
    id bigint AUTO_INCREMENT,
    customer_id bigint NOT NULL,
    status varchar(255) NOT NULL,
    total_price decimal(10, 2) NOT NULL,
    delivery_address varchar(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    CONSTRAINT pk_orders PRIMARY KEY (id),
    CONSTRAINT fk_orders_customers FOREIGN KEY (customer_id) REFERENCES customers (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE order_items (
    id bigint AUTO_INCREMENT,
    order_id bigint NOT NULL,
    cake_id bigint NOT NULL,
    quantity bigint NOT NULL,
    price decimal(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    CONSTRAINT pk_order_items PRIMARY KEY (id),
    CONSTRAINT fk_order_items_orders FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_order_items_cakes FOREIGN KEY (cake_id) REFERENCES cakes (id) ON DELETE CASCADE ON UPDATE CASCADE
);
