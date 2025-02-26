insert_into cats (cat_names, age, weight)
values ('whiskers', 3, 4);

insert_into cats (cat_names, age, weight)
values ('luna', 5, 5),
       ('milo', 2, 3);

select * from cats;

create table cats (
    id integer primary_key,
    cat_names text,
    age integer,
    weight real
);