create table zones
(
 name     varchar(20) not null
  constraint zones_pk
   primary key,
 geometry geometry(Polygon, 4326)
);

create table tracks
(
 id         integer
  constraint tracks_pk
   primary key,
 name       varchar(250)   not null,
 time       timestamp not null,
 coordinate geometry(Point, 4326)
);

create index tracks_time_index
 on tracks (time);

