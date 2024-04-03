alter table vessels
drop column is_deleted;

drop index vessels_pk;

alter table vessels
drop constraint vessels_pk;
