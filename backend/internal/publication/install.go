package publication

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"

	"216chan/backend/internal/model"
	"github.com/lib/pq"
)

//go:embed schema.sql
var Schema string

func Install(ctx context.Context, db *sql.DB, reader, writer, staff string) error {
	if reader == "" || writer == "" || reader == writer || staff == "" || staff == reader || staff == writer {
		return fmt.Errorf("distinct reader and writer roles required")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var database, owner string
	if err = tx.QueryRowContext(ctx, `SELECT current_database(),current_user`).Scan(&database, &owner); err != nil {
		return err
	}
	if reader == owner || writer == owner || staff == owner {
		return fmt.Errorf("request roles must differ from the installer")
	}
	var occupied bool
	if err = tx.QueryRowContext(ctx, `SELECT
 EXISTS(SELECT 1 FROM pg_namespace WHERE nspname NOT IN ('public','information_schema') AND nspname !~ '^pg_')
 OR EXISTS(SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public')
 OR EXISTS(SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname='public')
 OR EXISTS(SELECT 1 FROM pg_type t JOIN pg_namespace n ON n.oid=t.typnamespace WHERE n.nspname='public')`).Scan(&occupied); err != nil {
		return err
	}
	if occupied {
		return fmt.Errorf("installation requires an empty dedicated database; no migration is supported")
	}
	if _, err = tx.ExecContext(ctx, `REVOKE ALL ON DATABASE `+pq.QuoteIdentifier(database)+` FROM PUBLIC`); err != nil {
		return err
	}
	for _, role := range []string{reader, writer, staff} {
		if _, err = tx.ExecContext(ctx, `GRANT CONNECT ON DATABASE `+pq.QuoteIdentifier(database)+` TO `+pq.QuoteIdentifier(role)); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, Schema); err != nil {
		return err
	}
	if err = GrantRoles(ctx, tx, reader, writer, staff); err != nil {
		return err
	}
	// Ordinary commands preserve staff invariants and journal the initial decisions.
	if _, err = tx.ExecContext(ctx, `SELECT set_config('app.actor','install',true)`); err != nil {
		return err
	}
	for _, b := range Boards {
		if _, err = tx.ExecContext(ctx, `INSERT INTO pub.boards(slug,name,description,is_nsfw) VALUES($1,$2,$3,$4)`, b.Slug, b.Name, b.Description, b.IsNSFW); err != nil {
			return err
		}
		var post, conversation int64
		if err = tx.QueryRowContext(ctx, `SELECT pub.publish(gen_random_uuid(),$1,NULL,$2)`, b.Slug, boardRules[b.Slug]).Scan(&post); err != nil {
			return err
		}
		if err = tx.QueryRowContext(ctx, `SELECT conversation_id FROM pub.posts WHERE id=$1`, post).Scan(&conversation); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `SELECT pub.configure($1,0,'Board rules',NULL,true,true)`, conversation); err != nil {
			return err
		}
	}
	return tx.Commit()
}

var Boards = []model.Board{
	{Slug: "g", Name: "General", Description: "general talk, daily life, shitposting"},
	{Slug: "p", Name: "Programming", Description: "programming, dev tools, deploy pain"},
	{Slug: "v", Name: "Videogames", Description: "video games, reviews, game dev"},
	{Slug: "s", Name: "STEM", Description: "science, math, physics for nerds"},
	{Slug: "t", Name: "Technology", Description: "tech, hardware, PC builds"},
	{Slug: "a", Name: "Adult", Description: "18+ adult board", IsNSFW: true},
	{Slug: "b", Name: "Random", Description: "random, anything goes (within rules)", IsNSFW: true},
}

var boardRules = map[string]string{
	"g": `Welcome to /g/ - General. Tunisian imageboard for daily talk.

Rules for /g/:
1. Tunisian law applies. No CP, no gore with real people, no dox.
2. No spam or flood. No ads.
3. Keep it general. Use the right board for other topics.
4. Be civil. No targeted harassment.

This rules thread is pinned and locked.`,
	"p": `Welcome to /p/ - Programming. Code, tools, and work talk.

Rules for /p/:
1. Post code as text, not screenshots. Use the right board.
2. No homework dumps without your own try. Show what you did.
3. No spam or ads. No crypto scam.
4. Tunisian law and global rules apply.`,
	"v": `Welcome to /v/ - Videogames. Games and dev talk.

Rules for /v/:
1. Games and dev only. No piracy links.
2. Mark spoilers.
3. No spam or ads.
4. Follow global rules.`,
	"s": `Welcome to /s/ - STEM. Science and math.

Rules for /s/:
1. Show work. No low effort ask.
2. No homework without try.
3. Keep it on topic.
4. Follow global rules.`,
	"t": `Welcome to /t/ - Technology. Hardware and gear talk.

Rules for /t/:
1. Builds need parts list. No shop ads.
2. Keep it tech. No spam.
3. No dox or personal info.
4. Follow global rules.`,
	"a": `Welcome to /a/ - Adult. 18+ only.

Rules for /a/:
1. 18+ only. No CP. Zero tolerance.
2. Consent required. No non consensual content.
3. No gore or dox.
4. Follow Tunisian law.`,
	"b": `Welcome to /b/ - Random. Anything legal goes.

Rules for /b/:
1. Expect NSFW. Still no CP, no gore with real people, no dox.
2. No spam or flood.
3. Tunisian law applies. Mods will remove bad posts.
4. This is the random board. Have fun.`,
}

func GrantRoles(ctx context.Context, tx *sql.Tx, reader, writer, staff string) error {
	for i, name := range [3]string{reader, writer, staff} {
		q := pq.QuoteIdentifier(name)
		views := make([]string, len(roles[i].views))
		for j, view := range roles[i].views {
			views[j] = "pub." + pq.QuoteIdentifier(view)
		}
		functions := make([]string, len(roles[i].functions))
		for j, function := range roles[i].functions {
			functions[j] = "pub." + function
		}
		for _, statement := range []string{
			`GRANT USAGE ON SCHEMA pub TO ` + q,
			`GRANT SELECT ON ` + strings.Join(views, ",") + ` TO ` + q,
			`GRANT EXECUTE ON FUNCTION ` + strings.Join(functions, ",") + ` TO ` + q,
		} {
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return err
			}
		}
	}
	return nil
}
