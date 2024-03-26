create table users
(
 id          bigserial primary key,
 login       varchar(250)              not null unique,
 created_at  timestamptz default now() not null,
 modified_at timestamptz default now() not null,
 hash        varchar(60)               not null,
 role        int                       not null,
 is_deleted  bool        default false not null
);

