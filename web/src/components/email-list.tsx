import { useState, useEffect, useRef } from "react";
import { useParams } from "react-router-dom";
import { api, type EmailItem, type PaginatedResponse } from "../lib/api";
import EmailCard from "./email-card";
import { Input } from "./ui/input";
import { Button } from "./ui/button";
import { Skeleton } from "./ui/skeleton";
import { Search, RefreshCw } from "lucide-react";
import { cn } from "../lib/utils";

type Tab = "inbox" | "sent" | "starred";

export default function EmailList({ mailboxId }: { mailboxId: string }) {
  const { emailId } = useParams();
  const [emails, setEmails] = useState<EmailItem[]>([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [tab, setTab] = useState<Tab>("inbox");
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
      params.set("direction", tab);
    }
    if (debouncedSearch) params.set("search", debouncedSearch);

    try {
      const data = await api<PaginatedResponse>(`/emails?${params}`);
      if (reset) {
        setEmails(data.emails);
        setPage(1);
      } else {
        setEmails((prev) => [...prev, ...data.emails]);
        setPage(pageNum);
      }
      setHasMore(data.hasMore);
    } catch {
      // ignore
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }

  useEffect(() => {
    fetchEmails(true);
  }, [mailboxId, tab, debouncedSearch]);

  useEffect(() => {
    if (tab !== "inbox" || debouncedSearch) return;
    const interval = setInterval(() => fetchEmails(true), 15000);
    return () => clearInterval(interval);
  }, [mailboxId, tab, debouncedSearch]);

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
    <div className="w-96 shrink-0 border-r flex flex-col h-full bg-card">
      <div className="p-3 border-b space-y-2">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search emails..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>
        <div className="flex items-center justify-between">
          <div className="flex gap-1">
            {tabs.map((t) => (
              <button
                key={t.key}
                onClick={() => setTab(t.key)}
                className={cn(
                  "px-3 py-1 text-sm rounded-md transition-colors",
                  tab === t.key
                    ? "bg-primary text-primary-foreground"
                    : "hover:bg-accent"
                )}
              >
                {t.label}
              </button>
            ))}
          </div>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => fetchEmails(true)}
            title="Refresh"
          >
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {loading ? (
          <div className="p-4 space-y-3">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex gap-3">
                <Skeleton className="h-8 w-8 rounded-full" />
                <div className="flex-1 space-y-2">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-3 w-1/2" />
                </div>
              </div>
            ))}
          </div>
        ) : emails.length === 0 ? (
          <div className="flex items-center justify-center h-full text-muted-foreground text-sm p-4 text-center">
            {debouncedSearch
              ? "No emails match your search."
              : tab === "starred"
              ? "No starred emails."
              : tab === "sent"
              ? "No sent emails yet."
              : "No emails yet."}
          </div>
        ) : (
          <div>
            {emails.map((email) => (
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
