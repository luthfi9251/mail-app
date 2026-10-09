import { Link } from "react-router-dom";
import { cn } from "../lib/utils";
import { Avatar, AvatarFallback } from "./ui/avatar";
import { Star } from "lucide-react";
import type { EmailItem } from "../lib/api";

export default function EmailCard({
  email,
  mailboxId,
  selected,
  onStarToggle,
}: {
  email: EmailItem;
  mailboxId: string;
  selected?: boolean;
  onStarToggle?: () => void;
}) {
  const initials = (email.direction === "outbound" ? email.to : email.from)
    .charAt(0)
    .toUpperCase();

  return (
    <Link
      to={`/dashboard/${mailboxId}/${email.id}`}
      className={cn(
        "group relative block border-b border-border/60 px-4 py-3.5 transition-colors hover:bg-card",
        email.read ? "bg-transparent" : "bg-card",
        selected && "bg-accent hover:bg-accent"
      )}
    >
      {!email.read && <span className="absolute left-1.5 top-1/2 h-2 w-2 -translate-y-1/2 rounded-full bg-primary" />}
      <div className="flex items-start gap-3">
        <Avatar className="h-9 w-9 shrink-0">
          <AvatarFallback>{initials}</AvatarFallback>
        </Avatar>
        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between gap-2">
            <span className={cn("text-[13px] truncate", email.read ? "text-muted-foreground" : "font-semibold text-foreground")}>
              {email.direction === "outbound" ? `To: ${email.to}` : email.from}
            </span>
            <span className="text-xs text-muted-foreground shrink-0">
              {formatDate(email.createdAt)}
            </span>
          </div>
          <p className={cn("text-sm truncate mt-0.5", email.read ? "text-foreground/80" : "font-semibold text-foreground")}>
            {email.subject}
          </p>
          <p className="text-xs text-muted-foreground truncate mt-0.5">
            {email.textBody || email.htmlBody?.replace(/<[^>]*>/g, "") || ""}
          </p>
        </div>
        <button
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            onStarToggle?.();
          }}
          className="shrink-0 mt-1 rounded p-0.5 transition-colors hover:bg-muted"
        >
          <Star
            className={cn(
              "h-4 w-4",
              email.starred
                ? "fill-amber-400 text-amber-400"
                : "text-muted-foreground opacity-0 group-hover:opacity-100"
            )}
          />
        </button>
      </div>
    </Link>
  );
}

function formatDate(ts: number) {
  const d = new Date(ts);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  if (diff < 86400000) return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  if (diff < 604800000) return d.toLocaleDateString([], { weekday: "short" });
  return d.toLocaleDateString([], { month: "short", day: "numeric" });
}
