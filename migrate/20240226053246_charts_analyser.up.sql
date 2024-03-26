create table control_log
(
 id          bigserial primary key,
 vessel_id   bigserial                              not null,
 vessel_name varchar(250)                           not null,
 timestamp   timestamp with time zone default now() not null,
 control     boolean                                not null,
 comment     text
);

create index monitor_log_vessel_id_index
 on control_log (vessel_id);