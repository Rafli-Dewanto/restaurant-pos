CREATE TABLE carts (
    id INT AUTO_INCREMENT,
    customer_id INT NOT NULL,
    cake_id INT NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    CONSTRAINT pk_cart PRIMARY KEY (id),
    CONSTRAINT fk_cart_customer FOREIGN KEY (customer_id) REFERENCES customers (id),
    CONSTRAINT fk_cart_cake FOREIGN KEY (cake_id) REFERENCES cakes (id)
);
