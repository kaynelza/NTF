-- +goose Up
create schema if not exists ntf;

create table if not exists ntf.notification (
    id uuid primary key default gen_random_uuid(),
    email varchar(255) unique,
    title varchar(255),
    body text,
    status bigint,
    send_at TIMESTAMP with time zone,
    attempts bigint,
    last_error text,
    created_at TIMESTAMP with time zone default now(),
    sent_at TIMESTAMP with time zone
);

-- +goose Down
drop schema if exists ntf;