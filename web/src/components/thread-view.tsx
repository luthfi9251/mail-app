import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api, type EmailItem, type Attachment } from "../lib/api";
import { Avatar, AvatarFallback } from "./ui/avatar";
import { Badge } from "./ui/badge";
import { Separator } from "./ui/separator";
import { ScrollArea } from "./ui/scroll-area";
import { Button } from "./ui/button";
import ComposeDialog from "./compose-dialog";
import { Skeleton } from "./ui/skeleton";
import { ArrowLeft, ArrowUp, Paperclip, Reply } from "lucide-react";

export default function ThreadView({
  mailboxId,
  emailId,
}: {
  mailboxId: string;
  emailId: string;
}) {
  const [emails, setEmails] = useState<EmailItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [replyOpen, setReplyOpen] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    setLoading(true);
    api<{ emails: EmailItem[] }>(`/emails/${emailId}/thread`)
      .then((data) => setEmails(data.emails))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [emailId]);

  if (loading) {
    return (
      <div className="p-6 space-y-4 flex-1">
        <Skeleton className="h-6 w-1/2" />
        <Skeleton className="h-4 w-1/4" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (emails.length === 0) return null;

  const lastEmail = emails[emails.length - 1];

  function sanitizeHtml(html: string): string {
    return html
      .replace(/<script[^>]*>[\s\S]*?<\/script>/gi, "")
      .replace(/<iframe[^>]*>[\s\S]*?<\/iframe>/gi, "")
      .replace(/\son\w+="[^"]*"/gi, "")
      .replace(/\son\w+='[^']*'/gi, "");
  }

  function formatSize(bytes: number): string {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
  }

  return (
    <div className="flex-1 flex flex-col h-full">
      <ScrollArea className="flex-1">
        <div className="p-6">
          <Button
            variant="ghost"
            size="sm"
            className="lg:hidden mb-4 -ml-2 gap-1"
            onClick={() => navigate(`/dashboard/${mailboxId}`)}
          >
            <ArrowLeft className="h-4 w-4" />
            Back
          </Button>
          <h2 className="text-xl font-semibold mb-1">{emails[0].subject}</h2>
          <p className="text-sm text-muted-foreground mb-6">
            {emails.length} message{emails.length > 1 ? "s" : ""}
          </p>

          <div className="space-y-6">
            {emails.map((email, i) => (
              <div key={email.id}>
                <div className="flex items-start gap-3">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback>
                      {email.from.charAt(0).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-sm">{email.from}</span>
                        <span className="text-xs text-muted-foreground">to {email.to}</span>
                        {email.direction === "outbound" && (
                          <Badge variant="secondary" className="gap-1">
                            <ArrowUp className="h-3 w-3" />
                            Sent
                          </Badge>
                        )}
                      </div>
                      <span className="text-xs text-muted-foreground shrink-0">
                        {new Date(email.createdAt).toLocaleString()}
                      </span>
                    </div>
                    <div
                      className="mt-2 text-sm prose prose-sm max-w-none"
                      dangerouslySetInnerHTML={{
                        __html: email.htmlBody
                          ? sanitizeHtml(email.htmlBody)
                          : email.textBody?.replace(/\n/g, "<br>") || "",
                      }}
                    />

                    {email.attachments && email.attachments.length > 0 && (
                      <div className="mt-3 space-y-1">
                        {email.attachments.map((att: Attachment) => (
                          <div key={att.id} className="flex items-center gap-2 text-sm text-muted-foreground">
                            <Paperclip className="h-3 w-3" />
                            <span>{att.filename}</span>
                            <span className="text-xs">({formatSize(att.size)})</span>
                            <Button
                              variant="link"
                              size="sm"
                              className="h-auto p-0 text-xs"
                              onClick={async () => {
                                try {
                                  const data = await api<{ download_url: string }>(
                                    `/emails/${email.id}/attachments/${att.id}`
                                  );
                                  window.open(data.download_url, "_blank");
                                } catch {}
                              }}
                            >
                              Download
                            </Button>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
                {i < emails.length - 1 && <Separator className="my-4" />}
              </div>
            ))}
          </div>
        </div>
      </ScrollArea>

      <div className="border-t p-4">
        <Button
          variant="outline"
          className="gap-2"
          onClick={() => setReplyOpen(true)}
        >
          <Reply className="h-4 w-4" />
          Reply
        </Button>
      </div>

      <ComposeDialog
        mailboxId={mailboxId}
        open={replyOpen}
        onOpenChange={setReplyOpen}
        initialTo={
          lastEmail.direction === "inbound" ? lastEmail.from : lastEmail.to
        }
        initialSubject={
          lastEmail.subject.startsWith("Re:")
            ? lastEmail.subject
            : `Re: ${lastEmail.subject}`
        }
        inReplyTo={lastEmail.messageId}
      />
    </div>
  );
}
