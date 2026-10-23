CREATE SCHEMA pub;
ALTER DEFAULT PRIVILEGES REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;

CREATE TABLE pub.boards (
    slug text PRIMARY KEY CHECK (slug ~ '^[a-z][a-z0-9_-]{0,15}$'),
    name text NOT NULL DEFAULT '', description text NOT NULL DEFAULT '',
    is_nsfw boolean NOT NULL DEFAULT false, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE pub.conversations (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    board text NOT NULL REFERENCES pub.boards,
    opening_post_id bigint NOT NULL,
    locked boolean NOT NULL DEFAULT false,
    pinned boolean NOT NULL DEFAULT false,
    revision bigint NOT NULL DEFAULT 0 CHECK (revision>=0)
);
CREATE FUNCTION pub.webp_size(m bytea) RETURNS integer[] LANGUAGE sql IMMUTABLE PARALLEL SAFE
SET search_path=pg_catalog AS $$
SELECT CASE WHEN octet_length(m)<30 OR substring(m FROM 1 FOR 4)<>'\x52494646'::bytea
                 OR substring(m FROM 9 FOR 4)<>'\x57454250'::bytea THEN NULL
       ELSE CASE substring(m FROM 13 FOR 4)
       WHEN '\x56503858'::bytea THEN ARRAY[1+get_byte(m,24)+(get_byte(m,25)<<8)+(get_byte(m,26)<<16),
                                         1+get_byte(m,27)+(get_byte(m,28)<<8)+(get_byte(m,29)<<16)]
       WHEN '\x56503820'::bytea THEN ARRAY[(get_byte(m,26)+(get_byte(m,27)<<8))&16383,
                                         (get_byte(m,28)+(get_byte(m,29)<<8))&16383]
       END END $$;
CREATE TABLE pub.posts (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    conversation_id bigint NOT NULL REFERENCES pub.conversations,
    body text NOT NULL CHECK (char_length(body)<=5000),
    country_code text CHECK (country_code ~ '^[A-Z]{2}$'),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    media bytea CHECK (octet_length(media) BETWEEN 1 AND 8388608),
    digest bytea GENERATED ALWAYS AS (sha256(media)) STORED,
    media_width integer GENERATED ALWAYS AS ((pub.webp_size(media))[1]) STORED,
    media_height integer GENERATED ALWAYS AS ((pub.webp_size(media))[2]) STORED,
    state text NOT NULL DEFAULT 'visible' CHECK (state IN ('visible','hidden','erased')),
    revision bigint NOT NULL DEFAULT 0 CHECK (revision>=0),
    CHECK (CASE WHEN state='erased' THEN body='' AND media IS NULL
                ELSE body<>'' OR media IS NOT NULL END),
    UNIQUE(conversation_id,id)
);
ALTER TABLE pub.posts ALTER COLUMN media SET STORAGE EXTERNAL;
ALTER TABLE pub.conversations ADD CONSTRAINT opening_post_belongs
    FOREIGN KEY(id,opening_post_id) REFERENCES pub.posts(conversation_id,id)
    DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE pub.receipts (
    token uuid PRIMARY KEY,
    fingerprint bytea NOT NULL,
    post_id bigint NOT NULL UNIQUE REFERENCES pub.posts
);
CREATE TABLE pub.reports (
    post_id bigint PRIMARY KEY REFERENCES pub.posts,
    closed boolean NOT NULL DEFAULT false,
    revision bigint NOT NULL DEFAULT 0 CHECK (revision>=0)
);
CREATE TABLE pub.actions (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    actor text NOT NULL,
    kind text NOT NULL CHECK(kind IN ('post','conversation','review')),
    target bigint NOT NULL,
    before_state jsonb NOT NULL,
    after_state jsonb NOT NULL,
    reason text NOT NULL CHECK(char_length(reason) BETWEEN 1 AND 200),
    at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE VIEW pub.visible_posts WITH (security_barrier=true) AS
SELECT p.id,p.conversation_id,c.board,p.body,p.created_at,p.media IS NOT NULL AS has_media,p.country_code,
       p.id=c.opening_post_id AS is_op,octet_length(p.media) AS media_size,p.media_width,p.media_height
FROM pub.posts p JOIN pub.conversations c ON c.id=p.conversation_id
JOIN pub.posts root ON root.id=c.opening_post_id
WHERE p.state='visible' AND root.state='visible';
CREATE VIEW pub.public_boards WITH (security_barrier=true) AS SELECT * FROM pub.boards;
CREATE VIEW pub.visible_conversations WITH (security_barrier=true) AS
SELECT c.id,c.board,c.opening_post_id,c.locked,c.pinned,
       count(*) FILTER (WHERE NOT p.is_op)::integer AS reply_count,
       max(p.created_at) AS activity,min(p.created_at) FILTER (WHERE p.is_op) AS created_at
FROM pub.conversations c JOIN pub.visible_posts p ON p.conversation_id=c.id
GROUP BY c.id;
CREATE INDEX posts_conversation_order ON pub.posts(conversation_id,created_at,id);
CREATE INDEX conversations_board ON pub.conversations(board);
CREATE VIEW pub.visible_media WITH (security_barrier=true) AS
SELECT v.id AS post_id,p.media AS bytes,p.digest
FROM pub.visible_posts v JOIN pub.posts p ON p.id=v.id WHERE p.media IS NOT NULL;
CREATE VIEW pub.review_queue WITH (security_barrier=true) AS
SELECT r.post_id,r.revision,p.state,p.body,c.board,
       EXISTS(SELECT 1 FROM pub.visible_posts v WHERE v.id=p.id) AS exposed
FROM pub.reports r JOIN pub.posts p ON p.id=r.post_id
JOIN pub.conversations c ON c.id=p.conversation_id WHERE NOT r.closed;

CREATE FUNCTION pub.publish(p_token uuid,p_board text,p_conversation bigint,
                            p_body text,p_media bytea DEFAULT NULL,
                            p_source_digest bytea DEFAULT NULL,
                            p_country_code text DEFAULT NULL)
RETURNS bigint LANGUAGE plpgsql SECURITY DEFINER
SET search_path=pg_catalog,pub,pg_temp AS $$
DECLARE fingerprint bytea; receipt pub.receipts%ROWTYPE;
        c pub.conversations%ROWTYPE; pid bigint; cid bigint;
BEGIN
    IF p_token IS NULL OR p_board IS NULL OR p_body IS NULL
       OR char_length(p_body)>5000 OR (p_body='' AND p_media IS NULL)
       OR (p_media IS NOT NULL AND octet_length(p_media) NOT BETWEEN 1 AND 8388608)
       OR ((p_media IS NULL)<>(p_source_digest IS NULL))
       OR (p_source_digest IS NOT NULL AND octet_length(p_source_digest)<>32)
       OR (p_country_code IS NOT NULL AND p_country_code !~ '^[A-Z]{2}$')
    THEN RAISE EXCEPTION 'invalid submission' USING ERRCODE='22023'; END IF;
    -- Fingerprint the trusted original input so conversion does not change receipt identity.
    fingerprint:=sha256(convert_to(jsonb_build_array(p_board,p_conversation,p_body,
                               encode(p_source_digest,'hex'))::text,'UTF8'));
    PERFORM pg_advisory_xact_lock(hashtextextended(p_token::text,0));
    SELECT * INTO receipt FROM pub.receipts WHERE token=p_token;
    -- Acceptance must survive erasure without retaining input identity.
    IF FOUND AND octet_length(receipt.fingerprint)=0 THEN RETURN NULL; END IF;
    IF FOUND THEN
        IF receipt.fingerprint<>fingerprint THEN
            RAISE EXCEPTION 'receipt mismatch' USING ERRCODE='P2001';
        END IF;
        RETURN (SELECT id FROM pub.visible_posts WHERE id=receipt.post_id);
    END IF;
    IF p_conversation IS NOT NULL THEN
        SELECT * INTO c FROM pub.conversations WHERE id=p_conversation FOR UPDATE;
        IF NOT FOUND OR c.board<>p_board THEN
            RAISE EXCEPTION 'conversation not found' USING ERRCODE='P0002';
        END IF;
        -- A withdrawn root is indistinguishable from absence; only a lock is public.
        IF NOT EXISTS (SELECT 1 FROM pub.posts WHERE id=c.opening_post_id AND state='visible')
        THEN RAISE EXCEPTION 'conversation not found' USING ERRCODE='P0002'; END IF;
        IF c.locked THEN RAISE EXCEPTION 'conversation unavailable' USING ERRCODE='P0001'; END IF;
        cid:=c.id;
    END IF;
    pid:=nextval(pg_get_serial_sequence('pub.posts','id'));
    IF p_conversation IS NULL THEN
        INSERT INTO pub.conversations(board,opening_post_id)
        VALUES(p_board,pid) RETURNING id INTO cid;
    END IF;
    INSERT INTO pub.posts(id,conversation_id,body,media,country_code) OVERRIDING SYSTEM VALUE
    VALUES(pid,cid,p_body,p_media,NULLIF(p_country_code,'XX'));
    INSERT INTO pub.receipts VALUES(p_token,fingerprint,pid);
    RETURN pid;
END $$;

CREATE FUNCTION pub.set_post(p_post bigint,p_expected bigint,p_state text,p_reason text)
RETURNS bigint LANGUAGE plpgsql SECURITY DEFINER
SET search_path=pg_catalog,pub,pg_temp AS $$
DECLARE p pub.posts%ROWTYPE; cid bigint;
BEGIN
    IF p_expected IS NULL OR p_state IS NULL OR p_state NOT IN ('visible','hidden','erased')
       OR p_reason IS NULL OR char_length(p_reason) NOT BETWEEN 1 AND 200
       OR coalesce(current_setting('app.actor',true),'')=''
    THEN RAISE EXCEPTION 'invalid decision' USING ERRCODE='22023'; END IF;
    SELECT conversation_id INTO cid FROM pub.posts WHERE id=p_post;
    PERFORM 1 FROM pub.conversations WHERE id=cid FOR UPDATE;
    SELECT * INTO p FROM pub.posts WHERE id=p_post FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION 'missing post' USING ERRCODE='P0002'; END IF;
    IF p.revision<>p_expected THEN RAISE EXCEPTION 'stale decision' USING ERRCODE='P0001'; END IF;
    IF p.state=p_state THEN RETURN p.revision; END IF;
    IF p.state='erased' THEN RAISE EXCEPTION 'erasure is terminal' USING ERRCODE='P0001'; END IF;
    UPDATE pub.posts SET state=p_state,revision=revision+1,
        body=CASE WHEN p_state='erased' THEN '' ELSE body END,
        country_code=CASE WHEN p_state='erased' THEN NULL ELSE country_code END,
        media=CASE WHEN p_state='erased' THEN NULL ELSE media END WHERE id=p_post;
    IF p_state='erased' THEN
        UPDATE pub.receipts SET fingerprint='' WHERE post_id=p_post;
    END IF;
    INSERT INTO pub.actions(actor,kind,target,before_state,after_state,reason)
    VALUES(current_setting('app.actor'),'post',p_post,jsonb_build_object('state',p.state,'revision',p.revision),
           jsonb_build_object('state',p_state,'revision',p.revision+1),p_reason);
    RETURN p.revision+1;
END $$;

-- Revisions prevent stale forms overwriting staff decisions; no-ops need no history.
CREATE FUNCTION pub.configure(p_conversation bigint,p_expected bigint,p_reason text,
                             p_board text DEFAULT NULL,p_locked boolean DEFAULT NULL,
                             p_pinned boolean DEFAULT NULL)
RETURNS bigint LANGUAGE plpgsql SECURITY DEFINER
SET search_path=pg_catalog,pub,pg_temp AS $$
DECLARE c pub.conversations%ROWTYPE; target_board text; target_locked boolean; target_pinned boolean;
BEGIN
    IF p_expected IS NULL OR p_reason IS NULL OR char_length(p_reason) NOT BETWEEN 1 AND 200
       OR coalesce(current_setting('app.actor',true),'')=''
    THEN RAISE EXCEPTION 'invalid decision' USING ERRCODE='22023'; END IF;
    SELECT * INTO c FROM pub.conversations WHERE id=p_conversation FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION 'missing conversation' USING ERRCODE='P0002'; END IF;
    IF c.revision<>p_expected THEN RAISE EXCEPTION 'stale decision' USING ERRCODE='P0001'; END IF;
    target_board:=coalesce(p_board,c.board);
    target_locked:=coalesce(p_locked,c.locked);
    target_pinned:=coalesce(p_pinned,c.pinned);
    IF (target_board,target_locked,target_pinned)=(c.board,c.locked,c.pinned) THEN RETURN c.revision; END IF;
    UPDATE pub.conversations SET board=target_board,locked=target_locked,pinned=target_pinned,
                                revision=revision+1 WHERE id=p_conversation;
    INSERT INTO pub.actions(actor,kind,target,before_state,after_state,reason)
    VALUES(current_setting('app.actor'),'conversation',p_conversation,
       jsonb_build_object('board',c.board,'locked',c.locked,'pinned',c.pinned,'revision',c.revision),
       jsonb_build_object('board',target_board,'locked',target_locked,'pinned',target_pinned,'revision',c.revision+1),p_reason);
    RETURN c.revision+1;
END $$;

CREATE FUNCTION pub.report(p_post bigint) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog,pub,pg_temp AS $$
DECLARE cid bigint;
BEGIN
    SELECT conversation_id INTO cid FROM pub.posts WHERE id=p_post;
    PERFORM 1 FROM pub.conversations WHERE id=cid FOR UPDATE;
    IF EXISTS(SELECT 1 FROM pub.visible_posts WHERE id=p_post) THEN
        INSERT INTO pub.reports(post_id) VALUES(p_post) ON CONFLICT DO NOTHING;
    END IF;
    -- Uniform acknowledgments avoid revealing moderation state.
END $$;

CREATE FUNCTION pub.close_report(p_post bigint,p_expected bigint,p_reason text) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog,pub,pg_temp AS $$
DECLARE r pub.reports%ROWTYPE;
BEGIN
    IF p_expected IS NULL OR p_reason IS NULL OR char_length(p_reason) NOT BETWEEN 1 AND 200
       OR coalesce(current_setting('app.actor',true),'')=''
    THEN RAISE EXCEPTION 'invalid decision' USING ERRCODE='22023'; END IF;
    SELECT * INTO r FROM pub.reports WHERE post_id=p_post FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION 'missing report' USING ERRCODE='P0002'; END IF;
    IF r.revision<>p_expected THEN RAISE EXCEPTION 'stale decision' USING ERRCODE='P0001'; END IF;
    IF r.closed THEN RETURN r.revision; END IF;
    UPDATE pub.reports SET closed=true,revision=revision+1 WHERE post_id=p_post;
    INSERT INTO pub.actions(actor,kind,target,before_state,after_state,reason)
    VALUES(current_setting('app.actor'),'review',p_post,jsonb_build_object('closed',false,'revision',r.revision),
           jsonb_build_object('closed',true,'revision',r.revision+1),p_reason);
    RETURN r.revision+1;
END $$;

REVOKE ALL ON SCHEMA pub FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA pub FROM PUBLIC;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA pub FROM PUBLIC;

CREATE VIEW pub.staff_posts WITH (security_barrier=true) AS
SELECT p.id,p.body,p.state,p.revision,p.media IS NOT NULL AS has_media,c.id AS conversation_id,c.board,c.locked,c.pinned,
       c.revision AS conversation_revision,EXISTS(SELECT 1 FROM pub.visible_posts v WHERE v.id=p.id) AS exposed
FROM pub.posts p JOIN pub.conversations c ON c.id=p.conversation_id;

CREATE VIEW pub.staff_media WITH (security_barrier=true) AS
SELECT id AS post_id,media AS bytes FROM pub.posts WHERE media IS NOT NULL;
