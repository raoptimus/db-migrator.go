
CREATE OR REPLACE FUNCTION test_index_update() RETURNS trigger AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        INSERT INTO test2
        VALUES (
                NEW.id,
                setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'B') ||
                setweight(to_tsvector('english', COALESCE(array_to_string(NEW.synonyms, ' '), '')), 'C'),
                now()
        );
    ELSEIF(TG_OP = 'UPDATE') THEN
        UPDATE test2 SET
            document = setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'B') ||
                       setweight(to_tsvector('english', COALESCE(array_to_string(NEW.synonyms, ' '), '')), 'C'),
            updated_at = now()
        WHERE id = NEW.id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql
;

CREATE TRIGGER test_index_update_trigger
    AFTER INSERT OR UPDATE ON test
    FOR EACH ROW EXECUTE PROCEDURE test_index_update()
;
