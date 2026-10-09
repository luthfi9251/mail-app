import { useState, useEffect, useRef } from "react";
import { useParams } from "react-router-dom";
import { api, type EmailItem, type PaginatedResponse } from "../lib/api";
import EmailCard from "./email-card";
import { Input } from "./ui/input";
import { Button } from "./ui/button";
import { Search, RefreshCw, Menu } from "lucide-react";
import { cn } from "../lib/utils";

type Tab = "inbox" | "sent" | "starred";

export default function EmailList({ mailboxId, onToggleSidebar }: { mailboxId: string; onToggleSidebar?: () => void }) {
  const { emailId } = useParams();
  const [emails, setEmails] = useState<EmailItem[]>([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [tab, setTab] = useState<Tab>("inbox");
  const [syncing, setSyncing] = useState(false);
  const sentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(timer);
  }, [search]);

  async function fetchEmails(reset: boolean) {
    if (reset) { setLoading(true); } else { setLoadingMore(true); }

    const pageNum = reset ? 1 : page + 1;
    const params = new URLSearchParams({
      mailboxId,
      page: String(pageNum),
      limit: "20",
      _: String(Date.now()),
    });
    if (tab === "starred") {
      params.set("starred", "true");
    } else {
      params.set("direction", tab === "sent" ? "outbound" : "inbound");
    }
    if (debouncedSearch) params.set("search", debouncedSearch);

    try {
      const data = await api<PaginatedResponse>(`/emails?${params}`);
      if (reset) {
        setEmails(data?.emails ?? []);
        setPage(1);
      } else {
        setEmails((prev) => [...(prev ?? []), ...(data?.emails ?? [])]);
        setPage(pageNum);
      }
      setHasMore(data?.hasMore ?? false);
    } catch {
      // ignore
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }

  // Pull new mail from Resend into the DB, then reload the list.
  async function syncAndRefresh() {
    setSyncing(true);
    try {
      await api(`/mailboxes/${mailboxId}/sync`, { method: "POST" });
    } catch {
      // fall through: still show what is already in the DB
    } finally {
      setSyncing(false);
    }
    await fetchEmails(true);
  }

  useEffect(() => {
    fetchEmails(true);
  }, [mailboxId, tab, debouncedSearch]);

  useEffect(() => {
    setEmails([]);
  }, [mailboxId]);

  // The server marks the thread read when it is opened; mirror that locally.
  useEffect(() => {
    if (!emailId) return;
    const opened = emails.find((e) => e.id === emailId);
    if (!opened || opened.read) return;
    setEmails((prev) =>
      prev.map((e) => (e.threadId === opened.threadId ? { ...e, read: true } : e))
    );
  }, [emailId, emails]);

  useEffect(() => {
    syncAndRefresh();
  }, [mailboxId]);

  useEffect(() => {
    if (!sentinelRef.current) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !loadingMore) {
          fetchEmails(false);
        }
      },
      { threshold: 0.1 }
    );
    observer.observe(sentinelRef.current);
    return () => observer.disconnect();
  }, [hasMore, loadingMore, page]);

  async function handleStarToggle(email: EmailItem) {
    try {
      await api(`/emails/${email.id}/star`, {
        method: "PATCH",
        body: JSON.stringify({ starred: !email.starred }),
      });
      setEmails((prev) =>
        prev.map((e) => (e.id === email.id ? { ...e, starred: !e.starred } : e))
      );
    } catch {}
  }

  const tabs: { key: Tab; label: string }[] = [
    { key: "inbox", label: "Inbox" },
    { key: "sent", label: "Sent" },
    { key: "starred", label: "Starred" },
  ];

  return (
    <div className="w-full lg:w-[420px] lg:min-w-[420px] lg:max-w-[420px] lg:shrink-0 border-r flex flex-col h-full min-w-0 bg-background">
      <div className="p-4 border-b bg-card space-y-3">
        <div className="flex items-center gap-2 lg:hidden">
          <Button variant="ghost" size="icon" onClick={onToggleSidebar}>
            <Menu className="h-5 w-5" />
          </Button>
          <span className="font-semibold tracking-tight">Email App</span>
        </div>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search emails..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9 bg-muted/60 border-transparent shadow-none focus-visible:bg-card"
          />
        </div>
        <div className="flex items-center justify-between">
          <div className="flex gap-1 rounded-lg bg-muted p-1">
            {tabs.map((t) => (
              <button
                key={t.key}
                onClick={() => setTab(t.key)}
                className={cn(
                  "px-3 py-1 text-sm rounded-md transition-all font-medium",
                  tab === t.key
                    ? "bg-card text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground"
                )}
              >
                {t.label}
              </button>
            ))}
          </div>
          <Button
            variant="ghost"
            size="icon"
            onClick={syncAndRefresh}
            disabled={syncing || loading}
            title="Refresh"
          >
            <RefreshCw className={cn("h-4 w-4", (syncing || loading) && "animate-spin")} />
          </Button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto scroll-thin">
        {(emails?.length ?? 0) === 0 ? (
          loading || syncing ? null : (
          <div className="flex items-center justify-center h-full text-muted-foreground text-sm p-8 text-center">
            {debouncedSearch
              ? "No emails match your search."
              : tab === "starred"
              ? "No starred emails."
              : tab === "sent"
              ? "No sent emails yet."
              : "No emails yet."}
          </div>
          )
        ) : (
          <div>
            {emails?.map((email) => (
              <EmailCard
                key={email.id}
                email={email}
                mailboxId={mailboxId}
                selected={email.id === emailId}
                onStarToggle={() => handleStarToggle(email)}
              />
            ))}
            {hasMore && (
              <div ref={sentinelRef} className="py-4 text-center text-sm text-muted-foreground">
                {loadingMore ? "Loading more..." : "Scroll for more"}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
