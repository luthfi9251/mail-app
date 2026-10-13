# Email App

A self-hosted, multi-user web mail client for your own domain. Send and receive email from shared mailboxes (`support@yourdomain.com`, `hello@yourdomain.com`, ...) through a clean inbox UI, powered by [Resend](https://resend.com).

You bring a domain and a Resend account. This app gives you the inbox, threads, replies, and user/mailbox administration.

## Features

**Mail**
- Receive and send email on your own domain through Resend
- Conversation threading (replies are grouped into one thread)
- Rich-text composer (bold, italic, lists) and one-click reply
- Sender name in outgoing mail, e.g. `Luthfi <support@yourdomain.com>`
- Inbox, Sent, and Starred views with search and infinite scroll
- Unread indicators, star/unstar, attachment listing and download
- Manual refresh that pulls new mail from Resend (the refresh icon spins while loading)
- Email HTML is rendered in a sandboxed iframe with scripts disabled

**Users and mailboxes**
- First-run setup screen creates the initial admin account
- Admin dashboard to create/delete users, promote/demote admins, and create/delete mailboxes
- Each user only sees the mailboxes assigned to them
- Users can change their own password

**UI**
- Responsive layout with a mobile drawer and single-pane views
- React 19, Tailwind CSS v4, Radix UI

## How it works

```
Browser (React SPA) ──► Go API (Gin) ──► MongoDB
                              │
                              └──► Resend API  (send mail / fetch received mail)
```

- **Sending:** the API calls Resend with the mailbox address as the sender.
- **Receiving:** when you open a mailbox or press refresh, the API pulls newly received emails for that address from Resend and stores them in MongoDB. A webhook handler (`POST /api/webhooks/resend`, Svix-signature verified) exists but its route is commented out in `server/main.go` until it is needed.
- In production the Go server serves the built frontend from `web/dist`, so you run a single process.

## Prerequisites

- [Go](https://go.dev/dl/) (the version in `server/go.mod`)
- [Node.js](https://nodejs.org/) 20+ and npm
- A MongoDB 6+ instance
- [Docker](https://docs.docker.com/get-docker/) (optional, to run the app in a container)
- A [Resend](https://resend.com) account and a domain you control

## Setup

### 1. Configure Resend

Everything below is done in the [Resend Email Dashboard](https://resend.com/overview):

1. **Add your domain.** Go to **Domains → Add Domain**, then add the DNS records Resend shows you (SPF, DKIM, and MX for receiving) at your DNS provider. Wait until the domain shows as **Verified**.
2. **Enable receiving.** Make sure inbound email is enabled for the domain so Resend accepts mail sent to `anything@yourdomain.com`. See the [Resend receiving docs](https://resend.com/docs/dashboard/receiving/introduction).
3. **Create an API key.** Go to **API Keys → Create API Key** with permission to send and read emails. Copy the key (it starts with `re_`); you will not be able to see it again.

> Until the domain is verified, Resend will refuse to send from your mailbox addresses, and inbound mail will not arrive.

### 2. Clone and configure

```bash
git clone <your-repo-url> email-app
cd email-app
cp server/.env.example server/.env
```

Edit `server/.env`:

| Variable                | Description                                                                                   |
| ----------------------- | --------------------------------------------------------------------------------------------- |
| `MONGODB_URI`           | MongoDB connection string (**required**)                                                      |
| `RESEND_API_KEY`        | API key from the Resend dashboard (**required**)                                              |
| `JWT_SECRET`            | Random string, at least 32 chars: `openssl rand -hex 32` (**required**; the server refuses to start otherwise) |
| `MAILBOX_DOMAIN`        | Your verified Resend domain, e.g. `yourdomain.com`, no `@` (**required**)                     |
| `SETUP_TOKEN`           | Secret you must type on the first-run screen to create the admin. Unset = setup disabled      |
| `RESEND_WEBHOOK_SECRET` | Webhook signing secret (`whsec_...`). Only used once the webhook route is enabled             |
| `COOKIE_SECURE`         | Default `true`. Set `false` only for local development over plain HTTP                        |
| `CORS_ORIGINS`          | Comma-separated origins allowed to call the API cross-origin. Usually empty                  |
| `TRUSTED_PROXIES`       | Comma-separated IPs/CIDRs of your reverse proxy, so rate limiting sees real client IPs        |
| `PORT`                  | API port, default `8080`                                                                      |
| `STATIC_DIR`            | Built frontend directory, default `../web/dist`                                               |

For local development over `http://localhost`, set `COOKIE_SECURE=false`.

### 3. MongoDB

Point `MONGODB_URI` at your MongoDB instance. Keep it reachable only from the machines that need it, never from the internet.

### 4. Install and run

```bash
cd web && npm install && cd ..
make dev
```

`make dev` starts the API on `http://localhost:8080` and the Vite dev server on **http://localhost:5173**. Open the latter.

### 5. Create your admin account

Set `SETUP_TOKEN` in `server/.env` (e.g. `openssl rand -hex 16`) and restart the server. On first launch the login page asks you to **create the admin account**: enter the setup token, a name, email, and a password of at least 8 characters. You can remove `SETUP_TOKEN` afterwards.

- If your email is on `MAILBOX_DOMAIN`, that address becomes your first mailbox.
- Otherwise a mailbox `<your-local-part>@<MAILBOX_DOMAIN>` is created for you.

From the **Admin** page you can then create more users and mailboxes. Creating a user whose email is on `MAILBOX_DOMAIN` also creates a mailbox for them automatically.

Send a message to your new address from another account, press the refresh icon, and it should appear.

## Production

### Docker (recommended)

The multi-stage `Dockerfile` builds the frontend and the Go binary into a small distroless image that runs as non-root. `docker-compose.yml` starts **only the app**; it uses your existing MongoDB.

```bash
cp docker.env.example docker.env   # fill in MONGODB_URI, RESEND_API_KEY, JWT_SECRET, MAILBOX_DOMAIN
docker compose up -d --build
```

The app is published on port **30213**. Inside the container `localhost` is the container itself, so `MONGODB_URI` must use an address the container can reach (e.g. the host IP). Point your HTTPS reverse proxy or tunnel at `http://<host>:30213`.

### Without Docker

```bash
make build        # builds web/dist and the server/email-app binary
cd server && ./email-app
```

The server serves the SPA from `../web/dist` relative to where it runs, so start it from the `server/` directory. Put it behind a reverse proxy (Caddy, nginx, ...) with HTTPS.

Before deploying:
- Serve over HTTPS and keep `COOKIE_SECURE=true` (the default); run with `GIN_MODE=release`.
- Use a strong, unique `JWT_SECRET` and MongoDB password. Never commit `.env` files (they are git-ignored).
- Set `TRUSTED_PROXIES` to your reverse proxy address so login rate limiting keys on real client IPs.
- Set `SETUP_TOKEN` only for the first-run setup, and remove it afterwards.
- Never expose MongoDB to the internet.

Security notes:
- Sessions are HttpOnly, `SameSite=Lax` cookies, re-checked against the database on every request. Deleting a user, changing a role, or changing a password takes effect immediately; a password change signs out all other sessions.
- State-changing API requests must be `application/json` (CSRF guard). Login, setup and password change are rate limited (in memory, per process).
- Email HTML is rendered in a sandboxed iframe without scripts or forms. Remote images stay blocked until the reader clicks "Show images".

## Importing from v1

`server/cmd/import-v1` copies users, mailboxes and mailbox access from the v1 app database (Better Auth schema, default `mailbox-email`) into this app. Emails are not copied.

```bash
cd server
go run ./cmd/import-v1            # dry run: shows what would be created
go run ./cmd/import-v1 -apply     # import
```

It is safe to re-run: existing users (by email), mailboxes (by address) and assignments are left unchanged. v1 users keep their passwords. Their scrypt hashes are verified on login and upgraded to bcrypt automatically. Use `-skip a@x.com,b@y.com` to leave accounts out.

## Make commands

| Command             | What it does                          |
| ------------------- | ------------------------------------- |
| `make dev`          | Run API and frontend dev servers      |
| `make dev-server`   | Run only the API                      |
| `make dev-web`      | Run only the frontend                 |
| `make build`        | Build frontend and server             |
| `make clean`        | Remove build output                   |

## Project structure

```
server/   Go API (Gin, MongoDB driver, Resend)
  handlers/    auth, emails, mailboxes, admin, Resend sync/webhook
  middleware/  JWT auth and admin checks
  models/      MongoDB document types
web/      React + Vite + Tailwind frontend
  src/pages/       Login, Dashboard, Inbox, Admin
  src/components/  Sidebar, email list, thread view, composer, admin tables
```

## Troubleshooting

- **Cannot send / "failed to send email":** the domain is not verified in Resend, or `RESEND_API_KEY` is wrong.
- **No incoming mail:** check the MX records and that receiving is enabled for the domain in the Resend dashboard, then press refresh.
- **Images in emails do not load:** some senders' image hosts block embedding on other sites (`Cross-Origin-Resource-Policy`). This is a limitation of the sender's host, not a bug.
- **Server cannot connect to the database:** confirm MongoDB is running and `MONGODB_URI` is correct. In Docker, check `docker logs email-app`.
