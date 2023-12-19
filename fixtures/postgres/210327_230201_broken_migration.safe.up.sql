insert into test (id, rank, title, synonyms, updated_at)
values
    (1, 0, 'some title 1', ARRAY['some'], now()),
    (2, 0, 'some title 2', ARRAY['some'], now())
;

insert into test (id, rank, title, synonyms, updated_at)
values (1, 0, 'some title 3', ARRAY['some'], now());

insert into test (id, rank, title, synonyms, updated_at)
values (3, 0, 'some title 3', ARRAY['some'], now());
