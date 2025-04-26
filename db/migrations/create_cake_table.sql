CREATE TABLE cakes (
    id bigint NOT NULL AUTO_INCREMENT,
    title varchar(100) NOT NULL,
    description text NOT NULL,
    price decimal(10, 2) NOT NULL,
    category varchar(255) NOT NULL,
    rating decimal(3, 1) NOT NULL,
    image varchar(255) NOT NULL,
    created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at datetime NULL DEFAULT NULL,
    CONSTRAINT pk_cakes PRIMARY KEY (id)
);
