insert into test (id, rank, title, synonyms, updated_at)
values
    (2, 0, 'some title 2', ARRAY['some'], now()),
    (3, 0, 'some title 3', ARRAY['some'], now())
;

insert into test (id, rank, title, synonyms, updated_at)
values (2, 0, 'some title 4', ARRAY['some'], now());
