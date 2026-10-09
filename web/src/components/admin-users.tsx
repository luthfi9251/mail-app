import { useState, useEffect } from "react";
import { api } from "../lib/api";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Badge } from "./ui/badge";
import { Avatar, AvatarFallback } from "./ui/avatar";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";
import { Card } from "./ui/card";
import { useConfirm } from "./confirm-dialog";
import { Plus, Trash2, ShieldCheck, ShieldOff, Users } from "lucide-react";

interface AdminUser {
  id: string;
  email: string;
  name: string;
  role: string;
  mailboxCount: number;
}

export default function AdminUsers() {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [createOpen, setCreateOpen] = useState(false);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  useEffect(() => {
    api<AdminUser[]>("/admin/users").then((list) => setUsers(list ?? []));
  }, []);

  const [error, setError] = useState("");
  const confirm = useConfirm();

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await api("/admin/users", {
        method: "POST",
        body: JSON.stringify({ name, email, password }),
      });
    } catch (err: any) {
      setError(err.message);
      return;
    }
    setCreateOpen(false);
    setName("");
    setEmail("");
    setPassword("");
    api<AdminUser[]>("/admin/users").then((list) => setUsers(list ?? []));
  }

  function handleDelete(u: AdminUser) {
    confirm({
      title: "Delete user?",
      description: (
        <>
          <strong className="text-foreground">{u.name}</strong> ({u.email}) will be deleted and lose access to all
          their mailboxes. The mailboxes and their emails are kept. This cannot be undone.
        </>
      ),
      confirmLabel: "Delete user",
      onConfirm: async () => {
        await api(`/admin/users/${u.id}`, { method: "DELETE" });
        setUsers((prev) => (prev ?? []).filter((x) => x.id !== u.id));
      },
    });
  }

  function handleRole(u: AdminUser) {
    const role = u.role === "admin" ? "user" : "admin";
    confirm({
      title: role === "admin" ? "Make admin?" : "Revoke admin?",
      description:
        role === "admin" ? (
          <>
            <strong className="text-foreground">{u.name}</strong> will be able to manage all users, mailboxes and
            mailbox access.
          </>
        ) : (
          <>
            <strong className="text-foreground">{u.name}</strong> will lose access to the admin page.
          </>
        ),
      confirmLabel: role === "admin" ? "Make admin" : "Revoke admin",
      destructive: role !== "admin",
      onConfirm: async () => {
        await api(`/admin/users/${u.id}/role`, {
          method: "PATCH",
          body: JSON.stringify({ role }),
        });
        setUsers((prev) => (prev ?? []).map((x) => (x.id === u.id ? { ...x, role } : x)));
      },
    });
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold">Users</h2>
          <p className="text-sm text-muted-foreground">{users?.length ?? 0} total</p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus /> Create User
        </Button>
      </div>

      <Card className="divide-y overflow-hidden">
        {(users?.length ?? 0) === 0 && (
          <div className="flex flex-col items-center gap-2 p-10 text-sm text-muted-foreground">
            <Users className="h-8 w-8 opacity-40" />
            No users yet.
          </div>
        )}
        {users?.map((u) => (
          <div key={u.id} className="flex flex-wrap items-center justify-between gap-3 px-4 py-3.5 transition-colors hover:bg-muted/40 sm:px-5">
            <div className="flex min-w-0 items-center gap-3">
              <Avatar className="h-10 w-10">
                <AvatarFallback className="text-sm">{u.name.charAt(0).toUpperCase()}</AvatarFallback>
              </Avatar>
              <div className="min-w-0">
                <p className="truncate text-sm font-medium">{u.name}</p>
                <p className="truncate text-xs text-muted-foreground">{u.email}</p>
              </div>
            </div>
            <div className="flex items-center gap-2 sm:gap-3">
              <Badge variant={u.role === "admin" ? "default" : "secondary"} className="capitalize">
                {u.role}
              </Badge>
              <span className="hidden text-xs text-muted-foreground sm:inline">{u.mailboxCount} mailboxes</span>
              <Button
                variant="ghost"
                size="icon"
                title={u.role === "admin" ? "Revoke admin" : "Make admin"}
                onClick={() => handleRole(u)}
              >
                {u.role === "admin" ? <ShieldOff /> : <ShieldCheck />}
              </Button>
              <Button
                variant="ghost"
                size="icon"
                title="Delete user"
                className="hover:bg-destructive/10 hover:text-destructive"
                onClick={() => handleDelete(u)}
              >
                <Trash2 />
              </Button>
            </div>
          </div>
        ))}
      </Card>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create User</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            {error && (
              <p className="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">{error}</p>
            )}
            <Input placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} required />
            <Input type="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} required />
            <Input type="password" placeholder="Password (min. 8 characters)" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} />
            <Button type="submit" className="w-full">Create</Button>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
