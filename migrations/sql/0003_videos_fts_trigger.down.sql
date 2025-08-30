SET search_path TO app,
    public;
DROP TRIGGER IF EXISTS videos_fts_trg ON app.videos;
DROP FUNCTION IF EXISTS app.update_videos_fts();