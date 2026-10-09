import { useState, useEffect } from "react";
import { api } from "../lib/api";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";
import { Card } from "./ui/card";
import MailboxAccessDialog from "./mailbox-access-dialog";
import { useConfirm } from "./confirm-dialog";
import { Plus, Trash2, Inbox, Users } from "lucide-react";

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
  const [accessFor, setAccessFor] = useState<AdminMailbox | null>(null);
  const [createError, setCreateError] = useState("");
  const confirm = useConfirm();

  function reload() {
    api<AdminMailbox[]>("/admin/mailboxes").then((list) => setMailboxes(list ?? []));
  }

  useEffect(() => {
    api<{ mailboxDomain: string }>("/admin/config")
      .then((c) => setDomain(c?.mailboxDomain ?? ""))
      .catch(() => {});
    reload();
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setCreateError("");
    try {
      await api("/admin/mailboxes", {
        method: "POST",
        body: JSON.stringify({ localPart }),
      });
    } catch (err: any) {
      setCreateError(err.message);
      return;
    }
    setCreateOpen(false);
    setLocalPart("");
    reload();
  }

  function handleDelete(mb: AdminMailbox) {
    confirm({
      title: "Delete mailbox?",
      description: (
        <>
          <strong className="text-foreground">{mb.emailAddress}</strong> will be deleted and all {mb.userCount} user(s)
          will lose access. Mailboxes that still contain emails cannot be deleted.
        </>
      ),
      confirmLabel: "Delete mailbox",
      onConfirm: async () => {
        await api(`/admin/mailboxes/${mb.id}`, { method: "DELETE" });
        setMailboxes((prev) => (prev ?? []).filter((m) => m.id !== mb.id));
      },
    });
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
                <button
                  type="button"
                  className="text-xs text-muted-foreground hover:text-foreground hover:underline"
                  onClick={() => setAccessFor(mb)}
                >
                  {mb.userCount} user(s)
                </button>
              </div>
            </div>
            <div className="flex shrink-0 items-center gap-1">
              <Button variant="outline" size="sm" className="gap-1.5" onClick={() => setAccessFor(mb)}>
                <Users className="h-4 w-4" />
                <span className="hidden sm:inline">Manage access</span>
              </Button>
              <Button
                variant="ghost"
                size="icon"
                title="Delete mailbox"
                className="hover:bg-destructive/10 hover:text-destructive"
                onClick={() => handleDelete(mb)}
              >
                <Trash2 />
              </Button>
            </div>
          </div>
        ))}
      </Card>

      <MailboxAccessDialog
        mailbox={accessFor}
        onOpenChange={(open) => !open && setAccessFor(null)}
        onChanged={reload}
      />

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Mailbox</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            {createError && (
              <p className="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">{createError}</p>
            )}
            <div className="flex min-w-0 items-center gap-2">
              <Input
                placeholder="local-part"
                value={localPart}
                onChange={(e) => setLocalPart(e.target.value)}
                required
                className="flex-1"
              />
              <span className="min-w-0 truncate text-sm text-muted-foreground">{domain ? `@${domain}` : ""}</span>
            </div>
            <Button type="submit" className="w-full">Create</Button>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
