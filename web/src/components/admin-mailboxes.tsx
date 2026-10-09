import { useState, useEffect } from "react";
import { api } from "../lib/api";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";
import { Card } from "./ui/card";
import { Plus, Trash2, Inbox } from "lucide-react";

interface AdminMailbox {
  id: string;
  emailAddress: string;
  userCount: number;
}

export default function AdminMailboxes() {
  const [mailboxes, setMailboxes] = useState<AdminMailbox[]>([]);
  const [createOpen, setCreateOpen] = useState(false);
  const [localPart, setLocalPart] = useState("");
  const [domain, setDomain] = useState("");

  useEffect(() => {
    api<{ mailboxDomain: string }>("/admin/config")
      .then((c) => setDomain(c?.mailboxDomain ?? ""))
      .catch(() => {});
    api<AdminMailbox[]>("/admin/mailboxes").then((list) => setMailboxes(list ?? []));
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    await api("/admin/mailboxes", {
      method: "POST",
      body: JSON.stringify({ localPart }),
    });
    setCreateOpen(false);
    setLocalPart("");
    api<AdminMailbox[]>("/admin/mailboxes").then((list) => setMailboxes(list ?? []));
  }

  async function handleDelete(id: string) {
    try {
      await api(`/admin/mailboxes/${id}`, { method: "DELETE" });
      setMailboxes((prev) => (prev ?? []).filter((m) => m.id !== id));
    } catch (err: any) {
      alert(err.message);
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold">Mailboxes</h2>
          <p className="text-sm text-muted-foreground">{mailboxes?.length ?? 0} total</p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus /> Create Mailbox
        </Button>
      </div>

      <Card className="divide-y overflow-hidden">
        {(mailboxes?.length ?? 0) === 0 && (
          <div className="flex flex-col items-center gap-2 p-10 text-sm text-muted-foreground">
            <Inbox className="h-8 w-8 opacity-40" />
            No mailboxes yet.
          </div>
        )}
        {mailboxes?.map((mb) => (
          <div key={mb.id} className="flex items-center justify-between gap-3 px-4 py-3.5 transition-colors hover:bg-muted/40 sm:px-5">
            <div className="flex min-w-0 items-center gap-3">
              <span className="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-accent text-accent-foreground">
                <Inbox className="h-4 w-4" />
              </span>
              <div className="min-w-0">
                <p className="truncate text-sm font-medium">{mb.emailAddress}</p>
                <p className="text-xs text-muted-foreground">{mb.userCount} user(s)</p>
              </div>
            </div>
            <Button
              variant="ghost"
              size="icon"
              title="Delete mailbox"
              className="hover:bg-destructive/10 hover:text-destructive"
              onClick={() => handleDelete(mb.id)}
            >
              <Trash2 />
            </Button>
          </div>
        ))}
      </Card>

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
              <span className="text-sm text-muted-foreground">{domain ? `@${domain}` : ""}</span>
            </div>
            <Button type="submit" className="w-full">Create</Button>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
