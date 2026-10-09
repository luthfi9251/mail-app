const BASE = "/api";

export async function api<T = unknown>(
  path: string,
  options?: RequestInit
): Promise<T> {
  // The server rejects non-JSON writes (CSRF guard), so always send the header.
  const res = await fetch(BASE + path, {
    ...options,
    credentials: "include",
    headers: { "Content-Type": "application/json", ...options?.headers },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "request failed" }));
    throw new Error(body.error || `HTTP ${res.status}`);
  }

  return res.json();
}

export interface User {
  id: string;
  email: string;
  name: string;
  role: string;
}

export interface Mailbox {
  id: string;
  emailAddress: string;
  role: string;
}

export interface EmailItem {
  id: string;
  mailboxId: string;
  from: string;
  to: string;
  subject: string;
  htmlBody: string;
  textBody: string;
  messageId: string;
  inReplyTo: string | null;
  references: string[];
  threadId: string;
  direction: "inbound" | "outbound";
  read: boolean;
  starred: boolean;
  attachments: Attachment[];
  createdAt: number;
}

export interface Attachment {
  id: string;
  filename: string;
  size: number;
  content_type: string;
}

export interface PaginatedResponse {
  emails: EmailItem[];
  hasMore: boolean;
  total: number;
  page: number;
  totalPages: number;
}
