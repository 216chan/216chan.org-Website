CREATE TABLE IF NOT EXISTS boards (
    slug VARCHAR(10) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_nsfw BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS threads (
    id SERIAL PRIMARY KEY,
    board_slug VARCHAR(10) NOT NULL REFERENCES boards(slug) ON DELETE CASCADE,
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    reply_count INTEGER NOT NULL DEFAULT 0,
    bumped_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_threads_pinned_per_board ON threads(board_slug) WHERE is_pinned = true;
CREATE INDEX IF NOT EXISTS idx_threads_board_bumped ON threads(board_slug, bumped_at DESC);
CREATE INDEX IF NOT EXISTS idx_threads_board_created ON threads(board_slug, created_at DESC);

CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    thread_id INTEGER NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    board_slug VARCHAR(10) NOT NULL REFERENCES boards(slug) ON DELETE CASCADE,
    is_op BOOLEAN NOT NULL DEFAULT FALSE,
    body TEXT NOT NULL DEFAULT '',
    image_url TEXT,
    image_name TEXT,
    file_size_bytes INTEGER,
    country_code VARCHAR(2) NOT NULL DEFAULT 'XX',
    ip_hash VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (is_op = FALSE OR image_url IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_posts_thread_created ON posts(thread_id, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_posts_board_created ON posts(board_slug, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_board_file_size ON posts(board_slug, file_size_bytes) WHERE file_size_bytes IS NOT NULL;

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    ip_hash VARCHAR(64) NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sessions_last_seen ON sessions(last_seen_at DESC);

INSERT INTO boards (slug, name, description, is_nsfw) VALUES
    ('g', 'General', 'general talk, daily life, shitposting', FALSE),
    ('p', 'Programming', 'programming, dev tools, deploy pain', FALSE),
    ('v', 'Videogames', 'video games, reviews, game dev', FALSE),
    ('s', 'STEM', 'science, math, physics for nerds', FALSE),
    ('t', 'Technology', 'tech, hardware, PC builds', FALSE),
    ('a', 'Adult', '18+ adult board', TRUE),
    ('b', 'Random', 'random, anything goes (within rules)', TRUE)
ON CONFLICT (slug) DO NOTHING;

DO $$
DECLARE
    g_thread_id INTEGER;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM threads WHERE board_slug='g' AND is_pinned=true) THEN
        INSERT INTO threads (board_slug, is_pinned, bumped_at) VALUES ('g', true, NOW()) RETURNING id INTO g_thread_id;
        INSERT INTO posts (thread_id, board_slug, is_op, body, image_url, image_name, file_size_bytes, country_code) VALUES (g_thread_id, 'g', true, 'Welcome to /g/ - General. Read the rules before posting. This thread is pinned.', '/icons/banner.png', 'banner.png', 0, 'TN');
    END IF;
END $$;

DO $$
DECLARE
    b RECORD;
    tid INTEGER;
BEGIN
    FOR b IN SELECT slug FROM boards LOOP
        IF NOT EXISTS (SELECT 1 FROM threads WHERE board_slug=b.slug AND is_pinned=true) THEN
            INSERT INTO threads (board_slug, is_pinned, bumped_at) VALUES (b.slug, true, NOW()) RETURNING id INTO tid;
            INSERT INTO posts (thread_id, board_slug, is_op, body, image_url, image_name, file_size_bytes, country_code) VALUES (tid, b.slug, true, 'Welcome to /' || b.slug || '/ - This is the pinned thread.', '/icons/banner.png', 'banner.png', 0, 'TN');
        END IF;
    END LOOP;
END $$;