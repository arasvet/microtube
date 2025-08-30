SET search_path TO app,
    public;
CREATE OR REPLACE FUNCTION app.update_videos_fts() RETURNS trigger AS $$ BEGIN NEW.fts_tsv := setweight(
        to_tsvector(
            'simple',
            immutable_unaccent(coalesce(NEW.title, ''))
        ),
        'A'
    ) || setweight(
        to_tsvector(
            'simple',
            immutable_unaccent(coalesce(NEW.description, ''))
        ),
        'B'
    );
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS videos_fts_trg ON app.videos;
CREATE TRIGGER videos_fts_trg BEFORE
INSERT
    OR
UPDATE OF title,
    description ON app.videos FOR EACH ROW EXECUTE FUNCTION app.update_videos_fts();