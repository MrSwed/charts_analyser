alter table vessels
 add is_deleted bool default false not null;

alter table vessels
 add constraint vessels_pk
  unique (name);

