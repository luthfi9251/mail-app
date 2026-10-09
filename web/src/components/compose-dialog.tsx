import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../lib/api";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";
import { Input } from "./ui/input";
import { Button } from "./ui/button";
import RichTextEditor from "./rich-text-editor";
import { Send } from "lucide-react";

export default function ComposeDialog({
  mailboxId,
  open,
  onOpenChange,
  initialTo,
  initialSubject,
  inReplyTo,
}: {
  mailboxId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initialTo?: string;
  initialSubject?: string;
  inReplyTo?: string;
}) {
  const navigate = useNavigate();
  const [to, setTo] = useState(initialTo || "");
  const [subject, setSubject] = useState(initialSubject || "");
  const [body, setBody] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await api("/emails/send", {
        method: "POST",
        body: JSON.stringify({
          mailboxId,
          to,
          subject,
          htmlBody: body,
          textBody: body.replace(/<[^>]*>/g, ""),
          ...(inReplyTo ? { inReplyTo } : {}),
        }),
      });
      onOpenChange(false);
      navigate(`/dashboard/${mailboxId}`);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>{inReplyTo ? "Reply" : "New Email"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <p className="text-sm text-destructive bg-destructive/10 p-2 rounded">{error}</p>
          )}
          <Input
            type="email"
            placeholder="To"
            value={to}
            onChange={(e) => setTo(e.target.value)}
            required
          />
          <Input
            type="text"
            placeholder="Subject"
            value={subject}
            onChange={(e) => setSubject(e.target.value)}
            required
          />
          <RichTextEditor value={body} onChange={setBody} placeholder="Write your message..." />
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading} className="gap-2">
              <Send className="h-4 w-4" />
              {loading ? "Sending..." : "Send"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
