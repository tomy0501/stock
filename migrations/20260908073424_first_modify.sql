-- +goose Up
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS stock_room (
    facility_id BINARY(16) NOT NULL ,
    room_type_id BINARY(16) NOT NULL ,
    stock_date DATETIME NOT NULL ,
    stock_quantity INT NOT NULL ,
    reservation_quantity INT NOT NULL ,
    stop_sale_quantity INT NOT NULL ,
    PRIMARY KEY (facility_id, room_type_id, stock_date)
);

-- +goose Down
SELECT 'down SQL query';
