alter table control_log
 add column vessel_name varchar(250) not null;

alter table tracks
 add column vessel_name varchar(250) not null;

