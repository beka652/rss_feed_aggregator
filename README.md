# Gator - RSS Feed Aggregator

Gator is a command-line RSS feed aggregator written in Go. It lets multiple
users register, add and follow RSS feeds, periodically scrape those feeds for
new posts, and browse the posts that have been collected — all backed by a
Postgres database.

This project is based on the [boot.dev](https://www.boot.dev) RSS Feed
Aggregator guided project.

## Prerequisites

Before you can run this program, make sure you have the following installed:

- **[Go](https://go.dev/doc/install)** (version 1.27 or later)
- **[PostgreSQL](https://www.postgresql.org/download/)** (any recent version)

You'll also need the [`goose`](https://github.com/pressly/goose) migration
tool to set up the database schema:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## 1. Set up the database

Create a database for the project (call it whatever you like, e.g. `gator`):

```bash
createdb gator
```

Then run the migrations in `sql/schema` against it with `goose`:

```bash
cd sql/schema
goose postgres "postgres://<username>:<password>@localhost:5432/gator" up
```

## 2. Set up the config file

Gator reads its configuration from a `.gatorconfig.json` file in your **home
directory**. Create it there with the following contents, replacing the
connection string with your own Postgres credentials:

```json
{
  "db_url": "postgres://<username>:<password>@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

- `db_url` — the Postgres connection string Gator will use to connect to
  your database.
- `current_user_name` — leave this blank initially; it gets populated
  automatically once you register or log in as a user.

## 3. Install and run

From the project root, build the binary:

```bash
go build -o gator .
```

You can then run commands like this:

```bash
./gator <command> [arguments]
```

Alternatively, you can run it directly with `go run` during development:

```bash
go run . <command> [arguments]
```

## Commands

Here are some of the commands available:

| Command | Description |
| --- | --- |
| `register <name>` | Creates a new user and logs in as them. |
| `login <name>` | Logs in as an existing user. |
| `users` | Lists all registered users, marking the currently logged-in one. |
| `addfeed <name> <url>` | Adds a new RSS feed and automatically follows it. |
| `feeds` | Lists all feeds that have been added, along with who added them. |
| `follow <url>` | Follows an existing feed by its URL. |
| `following` | Lists the feeds the current user is following. |
| `unfollow <url>` | Unfollows a feed by its URL. |
| `agg` | Continuously scrapes the oldest-fetched feed for new posts (runs every 10 seconds until stopped). |
| `browse [limit]` | Displays recent posts from feeds the current user follows (defaults to 2 if no limit is given). |
| `reset` | Wipes the database (useful for starting fresh during development). |

### Example workflow

```bash
./gator register alice
./gator addfeed "Hacker News" "https://news.ycombinator.com/rss"
./gator agg          # run in one terminal to keep scraping for new posts
./gator browse 5      # in another terminal, view the 5 most recent posts
```
