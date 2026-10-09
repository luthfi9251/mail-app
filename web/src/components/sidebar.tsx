import { useEffect, useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import { api, type Mailbox, type User } from "../lib/api";
import { Button } from "./ui/button";
import { ScrollArea } from "./ui/scroll-area";
import { Separator } from "./ui/separator";
import { Avatar, AvatarFallback } from "./ui/avatar";
import LogoutButton from "./logout-button";
import ChangePasswordDialog from "./change-password-dialog";
import { Mail, Plus, Settings, X, KeyRound } from "lucide-react";

export default function Sidebar({ onCompose, onClose }: { onCompose: () => void; onClose?: () => void }) {
  const { mailboxId } = useParams();
  const navigate = useNavigate();
  const [mailboxes, setMailboxes] = useState<Mailbox[]>([]);
  const [user, setUser] = useState<User | null>(null);
  const [passwordOpen, setPasswordOpen] = useState(false);

  useEffect(() => {
    api<Mailbox[]>("/mailboxes").then(setMailboxes).catch(() => navigate("/login"));
    api<User>("/auth/me").then(setUser);
  }, []);

  return (
    <div className="w-64 min-w-[256px] max-w-[256px] border-r bg-card flex flex-col h-screen">
      <div className="p-4 flex items-center justify-between">
        <Link to="/dashboard" className="flex items-center gap-2 font-bold text-lg">
          <Mail className="h-5 w-5 text-primary" />
          Email App
        </Link>
        {onClose && (
          <Button variant="ghost" size="icon" className="lg:hidden" onClick={onClose}>
            <X className="h-5 w-5" />
          </Button>
        )}
      </div>

      <div className="px-4 mb-2">
        <Button onClick={onCompose} className="w-full gap-2">
          <Plus className="h-4 w-4" />
          Compose
        </Button>
      </div>

      <Separator />

      <ScrollArea className="flex-1 px-3 py-2">
        <div className="space-y-0.5">
          {mailboxes.map((mb) => (
            <Link
              key={mb.id}
              to={`/dashboard/${mb.id}`}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-md text-sm transition-colors ${
                mailboxId === mb.id
                  ? "bg-accent text-accent-foreground font-medium"
                  : "hover:bg-accent/50"
              }`}
            >
              <Mail className="h-4 w-4 shrink-0 text-muted-foreground" />
              <span className="truncate">{mb.emailAddress}</span>
            </Link>
          ))}
        </div>
      </ScrollArea>

      <Separator />

      <div className="p-4 space-y-2">
        {user?.role === "admin" && (
          <Link
            to="/admin"
            className="flex items-center gap-2 px-3 py-2 rounded-md text-sm hover:bg-accent/50 transition-colors"
          >
            <Settings className="h-4 w-4" />
            Admin
          </Link>
        )}
        <div className="flex items-center gap-3 px-3 py-2">
          <Avatar className="h-8 w-8 shrink-0">
            <AvatarFallback>
              {user?.name?.charAt(0)?.toUpperCase() || "?"}
            </AvatarFallback>
          </Avatar>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium truncate">{user?.name}</p>
            <p className="text-xs text-muted-foreground truncate">{user?.email}</p>
          </div>
        </div>
        <button
          onClick={() => setPasswordOpen(true)}
          className="flex items-center gap-2 w-full px-3 py-2 rounded-md text-sm text-muted-foreground hover:text-foreground hover:bg-accent/50 transition-colors"
        >
          <KeyRound className="h-4 w-4" />
          Change Password
        </button>
        <LogoutButton />
      </div>
      <ChangePasswordDialog open={passwordOpen} onOpenChange={setPasswordOpen} />
    </div>
  );
}
