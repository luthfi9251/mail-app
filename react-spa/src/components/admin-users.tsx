import { useState, useEffect } from "react";
import { api } from "../lib/api";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Badge } from "./ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Plus, Trash2, Shield, ShieldOff } from "lucide-react";

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
    api<AdminUser[]>("/admin/users").then(setUsers);
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    await api("/admin/users", {
      method: "POST",
      body: JSON.stringify({ name, email, password }),
    });
    setCreateOpen(false);
    setName("");
    setEmail("");
    setPassword("");
    api<AdminUser[]>("/admin/users").then(setUsers);
  }

  async function handleDelete(id: string) {
    await api(`/admin/users/${id}`, { method: "DELETE" });
    setUsers((prev) => prev.filter((u) => u.id !== id));
  }

  async function handleRole(id: string, role: string) {
    await api(`/admin/users/${id}/role`, {
      method: "PATCH",
      body: JSON.stringify({ role }),
    });
    setUsers((prev) => prev.map((u) => (u.id === id ? { ...u, role } : u)));
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Users ({users.length})</CardTitle>
        <Button size="sm" onClick={() => setCreateOpen(true)} className="gap-1">
          <Plus className="h-4 w-4" /> Create User
        </Button>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {users.map((u) => (
            <div key={u.id} className="flex items-center justify-between p-3 rounded-lg border">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-sm font-medium">
                  {u.name.charAt(0).toUpperCase()}
                </div>
                <div>
                  <p className="text-sm font-medium">{u.name}</p>
                  <p className="text-xs text-muted-foreground">{u.email}</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant={u.role === "admin" ? "default" : "secondary"}>
                  {u.role}
                </Badge>
                <span className="text-xs text-muted-foreground">{u.mailboxCount} mailboxes</span>
                {u.role === "admin" ? (
                  <Button variant="ghost" size="icon" onClick={() => handleRole(u.id, "user")}>
                    <ShieldOff className="h-4 w-4" />
                  </Button>
                ) : (
                  <Button variant="ghost" size="icon" onClick={() => handleRole(u.id, "admin")}>
                    <Shield className="h-4 w-4" />
                  </Button>
                )}
                <Button variant="ghost" size="icon" onClick={() => handleDelete(u.id)}>
                  <Trash2 className="h-4 w-4 text-destructive" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      </CardContent>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create User</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <Input placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} required />
            <Input type="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} required />
            <Input type="password" placeholder="Password" value={password} onChange={(e) => setPassword(e.target.value)} required />
            <Button type="submit" className="w-full">Create</Button>
          </form>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
