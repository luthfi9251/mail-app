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
        "group block border-b px-4 py-3 transition-colors hover:bg-accent/50",
        !email.read && "bg-primary/5",
        selected && "bg-accent"
      )}
    >
      <div className="flex items-start gap-3">
        <Avatar className="h-8 w-8 shrink-0 mt-0.5">
          <AvatarFallback>{initials}</AvatarFallback>
        </Avatar>
        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between gap-2">
            <span className={cn("text-sm truncate", !email.read && "font-semibold")}>
              {email.direction === "outbound" ? `To: ${email.to}` : email.from}
            </span>
            <span className="text-xs text-muted-foreground shrink-0">
              {formatDate(email.createdAt)}
            </span>
          </div>
          <p className={cn("text-sm truncate mt-0.5", !email.read && "font-medium")}>
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
          className="shrink-0 mt-1"
        >
          <Star
            className={cn(
              "h-4 w-4",
              email.starred
                ? "fill-yellow-400 text-yellow-400"
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
