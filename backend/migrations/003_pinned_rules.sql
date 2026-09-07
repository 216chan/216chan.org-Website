UPDATE posts SET body = 'Welcome to /g/ - General. Tunisian imageboard for daily talk.

Rules for /g/:
1. Tunisian law applies. No CP, no gore with real people, no dox.
2. No spam or flood. No ads.
3. Keep it general. Use the right board for other topics.
4. Be civil. No targeted harassment.

This thread is pinned and will be updated by mods.'
WHERE thread_id IN (SELECT id FROM threads WHERE board_slug='g' AND is_pinned=true) AND is_op=true;

UPDATE posts SET body = 'Welcome to /p/ - Programming. Code, tools, and work talk.

Rules for /p/:
1. Post code as text, not screenshots. Use the right board.
2. No homework dumps without your own try. Show what you did.
3. No spam or ads. No crypto scam.
4. Tunisian law and global rules apply.'

WHERE thread_id IN (SELECT id FROM threads WHERE board_slug='p' AND is_pinned=true) AND is_op=true;

UPDATE posts SET body = 'Welcome to /v/ - Videogames. Games and dev talk.

Rules for /v/:
1. Games and dev only. No piracy links.
2. Mark spoilers.
3. No spam or ads.
4. Follow global rules.'

WHERE thread_id IN (SELECT id FROM threads WHERE board_slug='v' AND is_pinned=true) AND is_op=true;

UPDATE posts SET body = 'Welcome to /s/ - STEM. Science and math.

Rules for /s/:
1. Show work. No low effort ask.
2. No homework without try.
3. Keep it on topic.
4. Follow global rules.'

WHERE thread_id IN (SELECT id FROM threads WHERE board_slug='s' AND is_pinned=true) AND is_op=true;

UPDATE posts SET body = 'Welcome to /t/ - Technology. Hardware and gear talk.

Rules for /t/:
1. Builds need parts list. No shop ads.
2. Keep it tech. No spam.
3. No dox or personal info.
4. Follow global rules.'

WHERE thread_id IN (SELECT id FROM threads WHERE board_slug='t' AND is_pinned=true) AND is_op=true;

UPDATE posts SET body = 'Welcome to /a/ - Adult. 18+ only.

Rules for /a/:
1. 18+ only. No CP. Zero tolerance.
2. Consent required. No non consensual content.
3. No gore or dox.
4. Follow Tunisian law.'

WHERE thread_id IN (SELECT id FROM threads WHERE board_slug='a' AND is_pinned=true) AND is_op=true;

UPDATE posts SET body = 'Welcome to /b/ - Random. Anything legal goes.

Rules for /b/:
1. Expect NSFW. Still no CP, no gore with real people, no dox.
2. No spam or flood.
3. Tunisian law applies. Mods will remove bad posts.
4. This is the random board. Have fun.'

WHERE thread_id IN (SELECT id FROM threads WHERE board_slug='b' AND is_pinned=true) AND is_op=true;

UPDATE posts SET body = 'Welcome to /' || board_slug || '/ - This is the pinned thread.

Rules:
1. No CP, no gore, no dox.
2. No spam.
3. Follow board topic and Tunisian law.'
WHERE is_op=true AND thread_id IN (SELECT id FROM threads WHERE is_pinned=true)
AND body LIKE 'Welcome to /%/ - This is the pinned thread.%';