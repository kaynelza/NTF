-- +goose Up
alter table if exists ntf.notification rename column email to recipient;


-- +goose Down
alter table if exists ntf.notification rename column recipient to email;