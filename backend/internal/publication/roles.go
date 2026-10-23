package publication

import (
	"context"
	"database/sql"
	"fmt"
	"slices"

	"github.com/lib/pq"
)

type role struct {
	kind             string
	views, functions []string
}

var publicViews = []string{"public_boards", "visible_posts", "visible_conversations", "visible_media"}

// Order is reader, writer, staff; CheckRoles and GrantRoles pair it with their arguments.
// Read-only: GrantRoles builds its own slices.
var roles = [3]role{
	{"reader", publicViews, []string{"report(bigint)"}},
	{"writer", publicViews, []string{"report(bigint)", "publish(uuid,text,bigint,text,bytea,bytea,text)"}},
	{"staff", slices.Concat(publicViews, []string{"staff_posts", "review_queue", "staff_media"}),
		[]string{"set_post(bigint,bigint,text,text)", "configure(bigint,bigint,text,text,boolean,boolean)", "close_report(bigint,bigint,text)"}},
}

func checkRole(ctx context.Context, db *sql.DB, r role) (string, error) {
	var name string
	var unsafe bool
	var missingViews, missingFunctions pq.StringArray
	err := db.QueryRowContext(ctx, `WITH allowed AS (SELECT $1::text[] AS views, ARRAY(SELECT 'pub.'||f FROM unnest($2::text[]) f) AS functions)
SELECT current_user,
 (SELECT rolsuper OR rolcreatedb OR rolcreaterole OR rolbypassrls OR rolreplication FROM pg_roles WHERE rolname=current_user)
 OR has_database_privilege(current_database(),'CREATE,TEMP')
 OR has_database_privilege(current_database(),'CONNECT WITH GRANT OPTION')
 OR has_schema_privilege('pub','USAGE WITH GRANT OPTION')
 OR EXISTS(SELECT 1 FROM pg_roles WHERE rolname<>current_user AND pg_has_role(current_user,oid,'SET'))
 OR EXISTS(SELECT 1 FROM pg_namespace WHERE nspname !~ '^pg_' AND nspname<>'information_schema' AND has_schema_privilege(oid,'CREATE'))
 OR NOT has_schema_privilege('pub','USAGE')
 OR EXISTS(SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='pub' AND
   CASE WHEN c.relkind='S' THEN has_sequence_privilege(c.oid,'USAGE,SELECT,UPDATE')
   ELSE has_table_privilege(c.oid,'INSERT,UPDATE,DELETE,TRUNCATE,REFERENCES,TRIGGER')
     OR has_table_privilege(c.oid,'SELECT WITH GRANT OPTION')
     OR has_any_column_privilege(c.oid,'INSERT,UPDATE,REFERENCES')
     OR has_any_column_privilege(c.oid,'SELECT WITH GRANT OPTION')
     OR (has_any_column_privilege(c.oid,'SELECT') AND NOT c.relname=ANY(allowed.views)) END)
 OR EXISTS(SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname='pub' AND
   (has_function_privilege(p.oid,'EXECUTE WITH GRANT OPTION') OR
    (has_function_privilege(p.oid,'EXECUTE') AND NOT (p.oid::regprocedure::text=ANY(allowed.functions))))),
 ARRAY(SELECT v FROM unnest(allowed.views) v WHERE NOT has_table_privilege('pub.'||v,'SELECT')),
 ARRAY(SELECT f FROM unnest(allowed.functions) f WHERE NOT has_function_privilege(f,'EXECUTE'))
FROM allowed`, pq.Array(r.views), pq.Array(r.functions)).Scan(&name, &unsafe, &missingViews, &missingFunctions)
	switch {
	case err != nil:
		return "", err
	case unsafe:
		return "", fmt.Errorf("unsafe %s role %s", r.kind, name)
	case len(missingViews) > 0:
		return "", fmt.Errorf("%s missing views %v", r.kind, missingViews)
	case len(missingFunctions) > 0:
		return "", fmt.Errorf("%s missing commands %v", r.kind, missingFunctions)
	}
	return name, nil
}
