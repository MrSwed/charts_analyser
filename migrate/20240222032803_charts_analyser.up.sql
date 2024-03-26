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
 id          bigserial primary key,
 vessel_id   bigint,
 vessel_name varchar(250)              not null,
 time        timestamptz default now() not null,
 location    geometry(Point, 4326)
);

create index tracks_time_index
 on tracks (time);

create index tracks_vessel_id_index
 on tracks (vessel_id);

create index tracks_location_index
 on tracks using gist (location);


create table vessels
(
 id         bigserial primary key,
 name       varchar(250),
 created_at timestamptz default now() not null
);
