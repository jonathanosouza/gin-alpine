-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    sp text := current_setting('search_path');
    tenant_schema text := btrim(split_part(sp, ',', 1));
BEGIN
    IF tenant_schema = '' OR tenant_schema = '$user' THEN
        tenant_schema := 'public';
    END IF;
    EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', tenant_schema);
    EXECUTE format('SET search_path = %I, public', tenant_schema);
END $$;

create type position_type as enum ('ADMIN', 'MANAGER', 'CUSTOMER', 'DEV');
create table if not exists roles(
    id serial primary key,
    role position_type not null default 'CUSTOMER',
    description varchar(100) not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    deleted_at timestamp
);
insert into roles (id, role, description)
values (
        1,
        'CUSTOMER',
        'cliente do sistema'
    );
insert into roles (id, role, description)
values (
        2,
        'MANAGER',
        'gerente do sistema'
    );
insert into roles (id, role, description)
values (
        3,
        'ADMIN',
        'administrador do sistema'
    );
insert into roles (id, role, description)
values (
        4,
        'DEV',
        'desenvolvedor do sistema'
    );
create table if not exists users(
    id serial primary key,
    uuid uuid not null,
    name varchar(50) not null,
    email varchar(100) not null,
    password varchar(255) not null,
    role_id int not null default 1,
    enabled boolean not null default true,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    deleted_at timestamp
);
ALTER TABLE users
ADD CONSTRAINT unique_user_email UNIQUE (email);
-- CONSTRAINTS
-- user -> roles
alter table if exists users
add constraint fk_users_role_id foreign key (role_id) references roles(id) on update cascade on delete cascade;
insert into users (uuid, name, email, "password", role_id)
values (
        '4a9b3fd5-6813-4c75-9598-5fd9ae202d88',
        'Admin',
        'admin@email.com',
        '$2a$10$zYC48a1doguo1VoCqbmQBezAUQJKVSbGnHgoPWInNFn2idbPABUoe',
        3
    ) on conflict (email) do nothing;
insert into users (uuid, name, email, "password", role_id)
values (
        '296446de-e045-4638-a4e5-a09e94136fee',
        'Jonas',
        'jonas.w.martins@gmail.com',
        '$2a$10$zYC48a1doguo1VoCqbmQBezAUQJKVSbGnHgoPWInNFn2idbPABUoe',
        4
    ) on conflict (email) do nothing;
create table if not exists links(
    id serial primary key,
    uuid uuid not null,
    data varchar(5000) not null,
    expires_at timestamp not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    deleted_at timestamp
);
CREATE INDEX idx_links_uuid ON links(uuid);
create type link_type as enum ('RESET_PASS', 'OTHER');
create table if not exists user_available_links(
    id serial primary key,
    user_id int not null,
    link_uuid uuid not null,
    type link_type not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    deleted_at timestamp
);
-- CONSTRAINTS
DROP INDEX IF EXISTS idx_links_uuid;
ALTER TABLE links
ADD CONSTRAINT idx_unique_links_uuid UNIQUE (uuid);
-- user_available_links -> users
alter table if exists user_available_links
add constraint fk_user_available_links_users_id foreign key (user_id) references users(id) on update cascade on delete cascade;
-- user_available_links -> links
alter table if exists user_available_links
add constraint fk_user_available_links_link_uuid foreign key (link_uuid) references links(uuid) on update cascade on delete cascade;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists user_available_links;
drop table IF EXISTS users;
drop table IF EXISTS roles;
drop table IF EXISTS links;
drop type IF EXISTS position_type;
drop type if exists link_type cascade;
DROP INDEX IF EXISTS idx_links_uuid;
DROP INDEX IF EXISTS idx_unique_links_uuid;
-- +goose StatementEnd
