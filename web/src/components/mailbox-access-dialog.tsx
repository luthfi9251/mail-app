import { useEffect, useState } from "react";
import { api } from "../lib/api";
import { Button } from "./ui/button";
import { Badge } from "./ui/badge";
import { Avatar, AvatarFallback } from "./ui/avatar";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";
import { useConfirm } from "./confirm-dialog";
import { Loader2, UserMinus, UserPlus } from "lucide-react";

interface Assignment {
  id: string;
  userId: string;
  name: string;
  email: string;
  role: string;
}

interface AdminUser {
  id: string;
  name: string;
  email: string;
}

const selectClass =
  "h-10 rounded-lg border border-input bg-card px-3 text-sm shadow-xs transition-all focus-visible:outline-none focus-visible:border-primary focus-visible:ring-4 focus-visible:ring-primary/15 disabled:cursor-not-allowed disabled:opacity-50";

export default function MailboxAccessDialog({
  mailbox,
  onOpenChange,
  onChanged,
}: {
  mailbox: { id: string; emailAddress: string } | null;
  onOpenChange: (open: boolean) => void;
  onChanged: () => void;
}) {
  const [assignments, setAssignments] = useState<Assignment[] | null>(null);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [userId, setUserId] = useState("");
  const [role, setRole] = useState("member");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const confirm = useConfirm();

  async function load(mailboxId: string) {
    const [list, allUsers] = await Promise.all([
      api<Assignment[]>(`/admin/mailboxes/${mailboxId}/assignments`),
      api<AdminUser[]>("/admin/users"),
    ]);
    setAssignments(list ?? []);
    setUsers(allUsers ?? []);
  }

  useEffect(() => {
    if (!mailbox) return;
    setAssignments(null);
    setUserId("");
    setRole("member");
    setError("");
    load(mailbox.id).catch((err) => setError(err.message));
  }, [mailbox]);

  const assigned = new Set((assignments ?? []).map((a) => a.userId));
  const available = users.filter((u) => !assigned.has(u.id));

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault();
    if (!mailbox || !userId) return;
    setBusy(true);
    setError("");
    try {
      await api("/admin/assignments", {
        method: "POST",
        body: JSON.stringify({ userId, mailboxId: mailbox.id, role }),
      });
      setUserId("");
      await load(mailbox.id);
      onChanged();
    } catch (err: any) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  function handleRemove(a: Assignment) {
    if (!mailbox) return;
    const mb = mailbox;
    confirm({
      title: "Remove access?",
      description: (
        <>
          <strong className="text-foreground">{a.name}</strong> ({a.email}) will no longer be able to read or send
          from <strong className="text-foreground">{mb.emailAddress}</strong>.
        </>
      ),
      confirmLabel: "Remove access",
      onConfirm: async () => {
        await api(`/admin/assignments/${a.id}`, { method: "DELETE" });
        await load(mb.id);
        onChanged();
      },
    });
  }

  return (
    <Dialog open={mailbox !== null} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Manage access</DialogTitle>
          <p className="break-all text-sm text-muted-foreground">{mailbox?.emailAddress}</p>
        </DialogHeader>

        {error && (
          <p className="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">{error}</p>
        )}

        <form onSubmit={handleAdd} className="flex flex-col gap-2 sm:flex-row">
          <select
            aria-label="User"
            className={`${selectClass} w-full min-w-0 sm:flex-1`}
            value={userId}
            onChange={(e) => setUserId(e.target.value)}
            disabled={assignments === null || available.length === 0}
            required
          >
            <option value="">{available.length === 0 ? "All users already have access" : "Select a user…"}</option>
            {available.map((u) => (
              <option key={u.id} value={u.id}>
                {u.name} ({u.email})
              </option>
            ))}
          </select>
          <div className="flex shrink-0 gap-2">
            <select aria-label="Role" className={`${selectClass} flex-1 sm:flex-none`} value={role} onChange={(e) => setRole(e.target.value)}>
              <option value="member">Member</option>
              <option value="owner">Owner</option>
            </select>
            <Button type="submit" disabled={busy || !userId} className="gap-2">
              {busy ? <Loader2 className="animate-spin" /> : <UserPlus />}
              Add
            </Button>
          </div>
        </form>

        <div className="max-h-80 divide-y overflow-y-auto rounded-lg border scroll-thin">
          {assignments === null && (
            <div className="flex justify-center p-6 text-muted-foreground">
              <Loader2 className="h-5 w-5 animate-spin" />
            </div>
          )}
          {assignments?.length === 0 && (
            <p className="p-6 text-center text-sm text-muted-foreground">No one has access to this mailbox yet.</p>
          )}
          {assignments?.map((a) => (
            <div key={a.id} className="flex items-center justify-between gap-3 px-3 py-2.5">
              <div className="flex min-w-0 items-center gap-3">
                <Avatar className="h-8 w-8">
                  <AvatarFallback className="text-xs">{a.name.charAt(0).toUpperCase()}</AvatarFallback>
                </Avatar>
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{a.name}</p>
                  <p className="truncate text-xs text-muted-foreground">{a.email}</p>
                </div>
              </div>
              <div className="flex shrink-0 items-center gap-1">
                <Badge variant={a.role === "owner" ? "default" : "secondary"} className="capitalize">
                  {a.role}
                </Badge>
                <Button
                  variant="ghost"
                  size="icon"
                  title="Remove access"
                  className="hover:bg-destructive/10 hover:text-destructive"
                  onClick={() => handleRemove(a)}
                >
                  <UserMinus />
                </Button>
              </div>
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
