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

create index tracks_coordinate_index
 on tracks using gist (coordinate);