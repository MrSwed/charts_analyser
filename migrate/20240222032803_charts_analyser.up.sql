create table zones
(
 name     varchar(20) not null
  constraint zones_pk
   primary key,
 geometry geometry(Polygon, 4326)
);

create table tracks
(
 id          SERIAL primary key,
 vessel_id   integer,
 vessel_name varchar(250) not null,
 time        timestamp    not null,
 coordinate  geometry(Point, 4326)
);

create index tracks_time_index
 on tracks (time);

create index tracks_vessel_id_index
 on tracks (vessel_id);

