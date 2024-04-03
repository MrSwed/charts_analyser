create table zones
(
 name     varchar(20) not null
  constraint zones_pk
   primary key,
 geometry geometry(Polygon, 4326)
);

create index zones_geometry_index
 on zones using gist (geometry);


create table tracks
(
 id        bigserial
  primary key,
 vessel_id bigint,
 time      timestamp with time zone default now() not null,
 location  geometry(Point, 4326)
);


create index tracks_time_index
 on tracks (time);

create index tracks_vessel_id_index
 on tracks (vessel_id);

create index tracks_location_index
 on tracks using gist (location);

create table vessels
(
 id         bigserial
  primary key,
 name       varchar(250)
  constraint vessels_pk
   unique,
 created_at timestamp with time zone default now() not null,
 is_deleted boolean                  default false not null
);


create table control_log
(
 id        bigserial
  primary key,
 vessel_id bigserial,
 timestamp timestamp with time zone default now() not null,
 control   boolean                                not null,
 comment   text
);


create index monitor_log_vessel_id_index
 on control_log (vessel_id);

create table control_dashboard
(
 vessel_id     bigint                not null
  primary key,
 state         boolean default false not null,
 timestamp     timestamp with time zone,
 control_start timestamp with time zone,
 control_end   timestamp with time zone,
 location      geometry(Point, 4326),
 current_zone  json
);


create table users
(
 id          bigserial
  primary key,
 login       varchar(250)                           not null
  unique,
 created_at  timestamp with time zone default now() not null,
 modified_at timestamp with time zone default now() not null,
 hash        varchar(60)                            not null,
 role        integer                                not null,
 is_deleted  boolean                  default false not null
);



