-- +goose Up
CREATE TABLE subscriptions(
    id SERIAL PRIMARY KEY,
    name varchar(50) NOT NULL,
    price INT NOT NULL,
    user_id UUID NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ
);

CREATE INDEX ON subscriptions (user_id, name, start_date, end_date);

-- +goose Down
DROP TABLE subscriptions;
