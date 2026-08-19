CREATE TABLE authors (
    id                TEXT PRIMARY KEY,
    handle            TEXT NOT NULL,
    handle_key        TEXT NOT NULL UNIQUE,
    display_name      TEXT NOT NULL,
    name_key          TEXT NOT NULL,
    avatar_source_url TEXT,
    avatar_path       TEXT,
    created_at        TEXT NOT NULL
);

CREATE INDEX authors_name_key_index ON authors(name_key);

CREATE TABLE author_identities (
    author_id TEXT NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
    kind      TEXT NOT NULL,
    value     TEXT NOT NULL,
    value_key TEXT NOT NULL,
    PRIMARY KEY (kind, value_key)
);

CREATE INDEX author_identities_author_index ON author_identities(author_id);

CREATE TABLE documents (
    id                   TEXT PRIMARY KEY,
    reading_kind         TEXT NOT NULL,
    acquisition_method   TEXT NOT NULL,
    original_format      TEXT NOT NULL,
    source_url           TEXT NOT NULL,
    title                TEXT NOT NULL,
    description          TEXT NOT NULL,
    site_name            TEXT NOT NULL,
    raw_byline           TEXT NOT NULL,
    author_id            TEXT REFERENCES authors(id) ON DELETE SET NULL,
    language             TEXT,
    published_at         TEXT,
    source_modified_at   TEXT,
    word_count           INTEGER NOT NULL CHECK (word_count >= 0),
    reading_time_minutes INTEGER NOT NULL CHECK (reading_time_minutes >= 0),
    thumbnail_path       TEXT,
    saved_at             TEXT NOT NULL
);

CREATE INDEX documents_author_index ON documents(author_id);
CREATE INDEX documents_saved_index ON documents(saved_at DESC, id);

