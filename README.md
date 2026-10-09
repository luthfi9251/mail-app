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
- **Receiving:** when you open a mailbox or press refresh, the API pulls newly received emails for that address from Resend and stores them in MongoDB. A webhook endpoint (`POST /api/webhooks/resend`) also exists.
- In production the Go server serves the built frontend from `web/dist`, so you run a single process.

## Prerequisites

- [Go](https://go.dev/dl/) (the version in `server/go.mod`)
- [Node.js](https://nodejs.org/) 20+ and npm
- [Docker](https://docs.docker.com/get-docker/) (optional, for a local MongoDB) or any MongoDB 6+ instance
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

| Variable         | Description                                                                 |
| ---------------- | --------------------------------------------------------------------------- |
| `MONGODB_URI`    | MongoDB connection string                                                   |
| `RESEND_API_KEY` | API key from the Resend dashboard                                           |
| `JWT_SECRET`     | Long random string used to sign login tokens (`openssl rand -hex 32`)       |
| `PORT`           | API port, default `8080`                                                    |
| `MAILBOX_DOMAIN` | Your verified Resend domain, e.g. `yourdomain.com` (no `@`)                 |

### 3. Start MongoDB

The included `docker-compose.yml` starts a local MongoDB that matches the default `MONGODB_URI`:

```bash
docker compose up -d
```

Already have MongoDB? Skip this and point `MONGODB_URI` at it.

### 4. Install and run

```bash
cd web && npm install && cd ..
make dev
```

`make dev` starts the API on `http://localhost:8080` and the Vite dev server on **http://localhost:5173**. Open the latter.

### 5. Create your admin account

On first launch the login page asks you to **create the admin account**. Enter a name, email, and password.

- If your email is on `MAILBOX_DOMAIN`, that address becomes your first mailbox.
- Otherwise a mailbox `<your-local-part>@<MAILBOX_DOMAIN>` is created for you.

From the **Admin** page you can then create more users and mailboxes. Creating a user whose email is on `MAILBOX_DOMAIN` also creates a mailbox for them automatically.

Send a message to your new address from another account, press the refresh icon, and it should appear.

## Production

```bash
make build        # builds web/dist and the server/email-app binary
cd server && ./email-app
```

The server serves the SPA from `../web/dist` relative to where it runs, so start it from the `server/` directory. Put it behind a reverse proxy (Caddy, nginx, ...) with HTTPS.

Before deploying:
- Use a strong, unique `JWT_SECRET`.
- Change the default MongoDB credentials in `docker-compose.yml` and `MONGODB_URI`.
- Update the CORS origin in `server/main.go` (hard-coded to `http://localhost:5173`, only needed if the frontend is served from a different origin).
- The login cookie is not marked `Secure`; enable that when serving over HTTPS.

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
- **Server cannot connect to the database:** confirm MongoDB is running and `MONGODB_URI` is correct (`docker compose ps`).
