# Email App Remake — Design Spec

**Date:** 2026-10-09
**Status:** Approved
**Stack:** React (Vite) + Golang (Gin) + MongoDB

## Goal

Remake the tripbooka email app with a new stack: React SPA frontend, Go/Gin REST API backend, MongoDB persistence. Same feature set: multi-user, multi-mailbox, send/reply via Resend, inbound webhook, admin panel, rich text compose, search, star, pagination.

## Architecture

```
react-spa/        → Vite + React + Tailwind + shadcn/ui
server/           → Go + Gin + mongo-go-driver + Resend SDK
```

Dev: `vite dev` on :5173, `go run` on :8080, MongoDB on :27017.
Prod: Go serves built React files from `react-spa/dist/`.

## Auth

- JWT-based (golang-jwt), bcrypt passwords
- Login: `POST /api/auth/login` → returns JWT in httpOnly cookie + JSON body
- Middleware extracts JWT, sets user info in Gin context
- Admin routes check `role == "admin"`

## Data Models (MongoDB)

### User
```
{
  _id: ObjectId,
  email: string (unique),
  name: string,
  password: string (bcrypt),
  role: "admin" | "user" (default "user"),
  createdAt: Date
}
```

### Mailbox
```
{
  _id: ObjectId,
  emailAddress: string (unique),
  resendDomainId: string,
  createdAt: Date
}
```

### UserMailbox
```
{
  _id: ObjectId,
  userId: ObjectId (ref User),
  mailboxId: ObjectId (ref Mailbox),
  role: "owner" | "admin" | "member",
  createdAt: Date
}
// unique index: { userId, mailboxId }
```

### Email
```
{
  _id: ObjectId,
  mailboxId: ObjectId (ref Mailbox),
  from: string,
  to: string,
  subject: string,
  htmlBody: string,
  textBody: string,
  messageId: string (unique),
  inReplyTo: string | null,
  references: string[],
  threadId: string (indexed),
  direction: "inbound" | "outbound",
  read: boolean,
  starred: boolean,
  attachments: [{ id, filename, size, content_type }],
  createdAt: Date
}
// indexes: { mailboxId, createdAt: -1 }, { mailboxId, direction, createdAt: -1 }
```

## API Endpoints

### Auth
| Method | Path | Description |
|--------|------|-------------|
| POST | /api/auth/login | Login, returns JWT |
| POST | /api/auth/logout | Clear cookie |

### Mailboxes
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/mailboxes | User's mailboxes |
| GET | /api/mailboxes/:id | Single mailbox info |

### Emails
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/emails | Paginated list (query: mailboxId, direction, page, limit, search, starred) |
| GET | /api/emails/:id/thread | Thread by threadId |
| PATCH | /api/emails/:id/star | Toggle star (body: { starred }) |
| POST | /api/emails/send | Send email via Resend |
| GET | /api/emails/:id/attachments/:attachmentId | Get download URL |

### Admin
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/admin/users | List users |
| POST | /api/admin/users | Create user |
| DELETE | /api/admin/users/:id | Delete user |
| PATCH | /api/admin/users/:id/role | Set role |
| GET | /api/admin/mailboxes | List mailboxes |
| POST | /api/admin/mailboxes | Create mailbox |
| DELETE | /api/admin/mailboxes/:id | Delete mailbox |
| POST | /api/admin/assignments | Assign mailbox to user |
| DELETE | /api/admin/assignments/:id | Unassign |

### Webhook
| Method | Path | Description |
|--------|------|-------------|
| POST | /api/webhooks/resend | Resend inbound email webhook |

## Frontend Routes

```
/login                    → Login page
/dashboard                → Redirect to first mailbox or empty state
/dashboard/:mailboxId     → 3-pane inbox
/dashboard/:mailboxId/:emailId → Thread preview (right pane)
/admin                    → Admin panel (users + mailboxes)
```

## Frontend Components

### Layout
- App shell with sidebar (mailbox list + compose button + profile)
- 3-pane: sidebar (w-64) | email list (w-96) | thread preview (flex-1)

### Shared (shadcn/ui)
- Button, Input, Card, Avatar, Badge, Dialog, Tabs, ScrollArea, Separator, Tooltip, Textarea, Skeleton

### Feature Components
- `Sidebar` — mailbox list, compose button, user profile, logout
- `EmailList` — self-fetching, tabs (Inbox/Sent/Starred), search, infinite scroll
- `EmailCard` — avatar, sender, subject, preview, star button, unread indicator
- `ThreadView` — thread messages with avatars, HTML rendering, attachments, reply form
- `ComposeDialog` — modal compose with Tiptap rich text editor
- `RichTextEditor` — Tiptap wrapper with toolbar (bold, italic, lists, link)
- `AdminUsers` — users table + CRUD
- `AdminMailboxes` — mailboxes table + CRUD

## Project Structure

```
email-app/
├── react-spa/
│   ├── src/
│   │   ├── components/
│   │   │   ├── ui/         (shadcn)
│   │   │   ├── sidebar.tsx
│   │   │   ├── email-list.tsx
│   │   │   ├── email-card.tsx
│   │   │   ├── thread-view.tsx
│   │   │   ├── compose-dialog.tsx
│   │   │   ├── rich-text-editor.tsx
│   │   │   ├── admin-users.tsx
│   │   │   ├── admin-mailboxes.tsx
│   │   │   ├── empty-state.tsx
│   │   │   ├── error-boundary-content.tsx
│   │   │   └── logout-button.tsx
│   │   ├── lib/
│   │   │   ├── api.ts      (fetch wrapper)
│   │   │   ├── auth.ts     (auth context)
│   │   │   └── utils.ts
│   │   ├── pages/
│   │   │   ├── LoginPage.tsx
│   │   │   ├── DashboardPage.tsx
│   │   │   ├── InboxPage.tsx
│   │   │   └── AdminPage.tsx
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── tailwind.config.ts
├── server/
│   ├── main.go
│   ├── go.mod
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── mailboxes.go
│   │   ├── emails.go
│   │   ├── admin.go
│   │   └── webhooks.go
│   ├── middleware/
│   │   └── auth.go
│   ├── models/
│   │   ├── user.go
│   │   ├── mailbox.go
│   │   ├── usermailbox.go
│   │   └── email.go
│   ├── database/
│   │   └── db.go
│   ├── .env
│   └── Dockerfile
├── docker-compose.yml
└── README.md
```

## Out of Scope

- Dark mode (CSS vars prepared, no toggle)
- Real-time push (polling via refresh)
- Email attachments outbound
- Email forwarding
- Bulk operations
