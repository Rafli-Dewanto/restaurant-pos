CREATE TABLE payments (
    `id` BIGINT,
    `order_id` BIGINT,
    `amount` decimal(10,2),
    `status` varchar(30),
    `payment_token` varchar(255),
    `payment_url` varchar(255),
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,

    CONSTRAINT pk_payments PRIMARY KEY (id),
    CONSTRAINT fk_payments_orders FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE ON UPDATE CASCADE
);
