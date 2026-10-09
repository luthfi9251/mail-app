import { useState, useEffect } from "react";
import { api } from "../lib/api";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Plus, Trash2 } from "lucide-react";

interface AdminMailbox {
  id: string;
  emailAddress: string;
  userCount: number;
}

export default function AdminMailboxes() {
  const [mailboxes, setMailboxes] = useState<AdminMailbox[]>([]);
  const [createOpen, setCreateOpen] = useState(false);
  const [localPart, setLocalPart] = useState("");

  useEffect(() => {
    api<AdminMailbox[]>("/admin/mailboxes").then(setMailboxes);
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    await api("/admin/mailboxes", {
      method: "POST",
      body: JSON.stringify({ localPart }),
    });
    setCreateOpen(false);
    setLocalPart("");
    api<AdminMailbox[]>("/admin/mailboxes").then(setMailboxes);
  }

  async function handleDelete(id: string) {
    try {
      await api(`/admin/mailboxes/${id}`, { method: "DELETE" });
      setMailboxes((prev) => prev.filter((m) => m.id !== id));
    } catch (err: any) {
      alert(err.message);
    }
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Mailboxes ({mailboxes.length})</CardTitle>
        <Button size="sm" onClick={() => setCreateOpen(true)} className="gap-1">
          <Plus className="h-4 w-4" /> Create Mailbox
        </Button>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {mailboxes.map((mb) => (
            <div key={mb.id} className="flex items-center justify-between p-3 rounded-lg border">
              <div>
                <p className="text-sm font-medium">{mb.emailAddress}</p>
                <p className="text-xs text-muted-foreground">{mb.userCount} user(s)</p>
              </div>
              <Button variant="ghost" size="icon" onClick={() => handleDelete(mb.id)}>
                <Trash2 className="h-4 w-4 text-destructive" />
              </Button>
            </div>
          ))}
        </div>
      </CardContent>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Mailbox</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="flex items-center gap-2">
              <Input
                placeholder="local-part"
                value={localPart}
                onChange={(e) => setLocalPart(e.target.value)}
                required
                className="flex-1"
              />
              <span className="text-sm text-muted-foreground">@domain.com</span>
            </div>
            <Button type="submit" className="w-full">Create</Button>
          </form>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
