import { useEffect, useState } from "react";
import { Link, useParams, useNavigate, useLocation } from "react-router-dom";
import { api, type Mailbox, type User } from "../lib/api";
import { cn } from "../lib/utils";
import { ScrollArea } from "./ui/scroll-area";
import { Avatar, AvatarFallback } from "./ui/avatar";
import LogoutButton from "./logout-button";
import ChangePasswordDialog from "./change-password-dialog";
import { Mail, Plus, Settings, X, KeyRound, Inbox } from "lucide-react";

export default function Sidebar({ onCompose, onClose }: { onCompose: () => void; onClose?: () => void }) {
  const { mailboxId } = useParams();
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const [mailboxes, setMailboxes] = useState<Mailbox[]>([]);
  const [user, setUser] = useState<User | null>(null);
  const [passwordOpen, setPasswordOpen] = useState(false);

  useEffect(() => {
    api<Mailbox[]>("/mailboxes").then((list) => setMailboxes(list ?? [])).catch(() => navigate("/login"));
    api<User>("/auth/me").then(setUser);
  }, []);

  const adminActive = pathname.startsWith("/admin");

  return (
    <div className="w-64 min-w-[256px] max-w-[256px] bg-sidebar text-sidebar-foreground flex flex-col h-dvh">
      <div className="px-5 h-16 flex items-center justify-between shrink-0">
        <Link to="/dashboard" className="flex items-center gap-2.5 font-semibold text-[17px] tracking-tight text-white">
          <span className="grid h-8 w-8 place-items-center rounded-lg bg-gradient-to-br from-indigo-400 to-violet-500 shadow-lg shadow-indigo-900/40">
            <Mail className="h-4 w-4 text-white" />
          </span>
          Email App
        </Link>
        {onClose && (
          <button
            onClick={onClose}
            className="lg:hidden rounded-lg p-1.5 text-sidebar-muted hover:bg-sidebar-accent hover:text-white transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        )}
      </div>

      <div className="px-4 pb-4 shrink-0">
        <button
          onClick={onCompose}
          className="flex w-full items-center justify-center gap-2 rounded-xl bg-white px-4 py-2.5 text-sm font-semibold text-slate-900 shadow-sm transition-all hover:bg-indigo-50 hover:shadow-md active:scale-[0.98]"
        >
          <Plus className="h-4 w-4" />
          Compose
        </button>
      </div>

      <p className="px-6 pb-2 text-[11px] font-semibold uppercase tracking-wider text-sidebar-muted/80">Mailboxes</p>

      <ScrollArea className="flex-1 px-3">
        <div className="space-y-0.5 pb-2">
          {mailboxes?.map((mb) => {
            const active = mailboxId === mb.id;
            return (
              <Link
                key={mb.id}
                to={`/dashboard/${mb.id}`}
                className={cn(
                  "relative flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm transition-colors",
                  active
                    ? "bg-sidebar-accent font-medium text-white"
                    : "text-sidebar-foreground/80 hover:bg-sidebar-accent/60 hover:text-white"
                )}
              >
                {active && <span className="absolute left-0 top-2 bottom-2 w-[3px] rounded-r-full bg-indigo-400" />}
                <Inbox className={cn("h-4 w-4 shrink-0", active ? "text-indigo-300" : "text-sidebar-muted")} />
                <span className="truncate">{mb.emailAddress}</span>
              </Link>
            );
          })}
        </div>
      </ScrollArea>

      <div className="shrink-0 border-t border-sidebar-border p-3 space-y-0.5">
        {user?.role === "admin" && (
          <Link
            to="/admin"
            className={cn(
              "flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors",
              adminActive
                ? "bg-sidebar-accent font-medium text-white"
                : "text-sidebar-foreground/80 hover:bg-sidebar-accent/60 hover:text-white"
            )}
          >
            <Settings className="h-4 w-4 text-sidebar-muted" />
            Admin
          </Link>
        )}
        <button
          onClick={() => setPasswordOpen(true)}
          className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-sidebar-foreground/80 transition-colors hover:bg-sidebar-accent/60 hover:text-white"
        >
          <KeyRound className="h-4 w-4 text-sidebar-muted" />
          Change Password
        </button>
        <LogoutButton />

        <div className="mt-2 flex items-center gap-3 rounded-xl bg-sidebar-accent/50 px-3 py-2.5">
          <Avatar className="h-9 w-9 shrink-0">
            <AvatarFallback className="text-sm">{user?.name?.charAt(0)?.toUpperCase() || "?"}</AvatarFallback>
          </Avatar>
          <div className="flex-1 min-w-0">
            <p className="truncate text-sm font-medium text-white">{user?.name}</p>
            <p className="truncate text-xs text-sidebar-muted">{user?.email}</p>
          </div>
        </div>
      </div>
      <ChangePasswordDialog open={passwordOpen} onOpenChange={setPasswordOpen} />
    </div>
  );
}
