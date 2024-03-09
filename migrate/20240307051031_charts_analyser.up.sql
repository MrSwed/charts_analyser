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