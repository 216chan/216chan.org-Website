# Deploy

Build with `docker build -t 216chan .` (listens on `0.0.0.0:8080`), or
`go build -o server ./cmd/server` on a host with bubblewrap and ImageMagick
(JPEG and WebP delegates) installed.

**1. Database.** Create a fresh database owned by a non-serving owner role, plus
three login roles:

```sql
CREATE ROLE chan_owner  LOGIN PASSWORD '...' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE chan_reader LOGIN PASSWORD '...' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE chan_writer LOGIN PASSWORD '...' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE ROLE chan_staff  LOGIN PASSWORD '...' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
CREATE DATABASE publication OWNER chan_owner;
```

**2. Install once.** The installer refuses a database that is not empty. Ordinary
startup never changes the schema. Never give the owner credentials to the
serving process.

```sh
PUBLICATION_OWNER_URL='postgres://chan_owner:...@host/publication?sslmode=require' \
PUBLICATION_READER_ROLE=chan_reader PUBLICATION_WRITER_ROLE=chan_writer \
PUBLICATION_STAFF_ROLE=chan_staff ./server -install
```

**3. Staff config.** Each operator gets a random secret; the config stores its
SHA-256 digest. Operators log in to `/staff` with HTTP Basic auth as
`name:secret`. All staff have the same authority.

```sh
openssl rand -base64 32 | tr '+/' '-_' | tr -d '='   # csrf_key
s=$(openssl rand -hex 32); echo "$s"; printf %s "$s" | sha256sum | cut -d' ' -f1   # secret, then digest
```

```json
{"origin":"https://example.org","csrf_key":"...","users":{"alice":"<digest>"},"tls_proxy":true}
```

Set `tls_proxy` only when a TLS-terminating proxy is the sole route to the
process. Without it, staff requires direct TLS or a loopback peer. A container's
published port is not a loopback peer, so `/staff` answers 403 there until a TLS
proxy fronts it.

**4. Serve.** `./server -check` validates the full configuration and exits.
`/healthz` is liveness; `/readyz` checks database connectivity.

| Variable | Default | Meaning |
|---|---|---|
| `PUBLICATION_DATABASE_URL` | required | `postgres://host:5432/publication?sslmode=require`, **without credentials**; `sslmode=disable` only on a private host or container network |
| `PUBLICATION_{READER,WRITER,STAFF}_ROLE` / `_PASSWORD` | required | the three request logins |
| `STAFF_CONFIG` or `STAFF_CONFIG_JSON` | required | path to a mode-0600 file, or the JSON inline (not both) |
| `PUBLICATION_LISTEN` | `127.0.0.1:8081` | listen address (the container sets `0.0.0.0:8080`) |
| `SITE_ORIGIN` | `https://216chan.org` | used in canonical and og: links |
| `IMAGE_SANDBOX` | `bwrap` | `none` drops namespace isolation for hosts without user namespaces and for unprivileged containers (rootless podman: `bwrap: Can't mount proc`), at the cost of exposing the server to decoder exploits |
| `POSTING_DISABLED` | `false` | emergency switch: posts get `503`; reading still works |
| `GEOIP_DB` | unset (flags show `XX`) | country `.mmdb`; a file that is set but unusable stops startup |
| `GEOIP_TRUSTED_PROXIES` | none | CIDRs whose single `X-Real-IP` header is trusted; other forwarding headers are ignored |
| `PUBLICATION_NETWORK_LANDLOCK` | `false` | limits the server to its HTTP and PostgreSQL ports (Landlock ABI 10+, needs a numeric DB host); extra defense, not a firewall |

**Country flags (optional).** Use your own GeoLite2-Country or GeoIP2-Country
`.mmdb` from MaxMind; other editions and vendors are refused, and its license
forbids publishing the file. Mount it read-only and restart after updates.
Behind a proxy, list the proxy's CIDRs in `GEOIP_TRUSTED_PROXIES` and make the
proxy **overwrite** `X-Real-IP` with the client address (nginx:
`proxy_set_header X-Real-IP $remote_addr;`), or posters can fake their country.

Logs are JSON on stderr. Client addresses, post bodies and receipts are never
logged. Images live in Postgres, so measure volume and backup growth.
