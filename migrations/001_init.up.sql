create schema if not exists pvz_service;

create table if not exists pvz_service.user (
    user_id uuid primary key default gen_random_uuid(),
    email varchar(255) unique not null,
    password varchar(255) not null,
    role varchar(20) not null check (role in ('employee', 'moderator'))
);

create index idx_user_email ON pvz_service.user(email);

create table if not exists pvz_service.pvz (
    pvz_id uuid primary key default gen_random_uuid(),
    registration_date date not null default current_date,
    city varchar(255) not null
);

create table if not exists pvz_service.reception (
    reception_id uuid primary key default gen_random_uuid(),
    started_at timestamp not null default current_timestamp,
    pvz_id uuid not null,
    status varchar(20) not null default 'in_progress' check (status in ('in_progress', 'close')),
    constraint fk_pvz_id foreign key (pvz_id) references pvz_service.pvz (pvz_id)
);

create table if not exists pvz_service.product (
    product_id uuid primary key default gen_random_uuid(),
    added_at timestamp not null default current_timestamp,
    product_type varchar(255) not null,
    reception_id uuid not null,
    constraint fk_reception_id foreign key (reception_id) references pvz_service.reception (reception_id)
);


