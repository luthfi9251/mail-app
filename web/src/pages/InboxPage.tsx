import { useState } from "react";
import { useParams } from "react-router-dom";
import Sidebar from "../components/sidebar";
import EmailList from "../components/email-list";
import ThreadView from "../components/thread-view";
import ComposeDialog from "../components/compose-dialog";
import EmptyState from "../components/empty-state";

export default function InboxPage() {
  const { mailboxId, emailId } = useParams();
  const [composeOpen, setComposeOpen] = useState(false);

  if (!mailboxId) return null;

  return (
    <div className="flex h-screen">
      <Sidebar onCompose={() => setComposeOpen(true)} />
      <EmailList mailboxId={mailboxId} />
      <div className="flex-1 flex flex-col h-full overflow-hidden">
        {emailId ? (
          <ThreadView mailboxId={mailboxId} emailId={emailId} />
        ) : (
          <EmptyState
            title="Select an email to read"
            description="Choose an email from the list to view its contents."
          />
        )}
      </div>
      <ComposeDialog
        mailboxId={mailboxId}
        open={composeOpen}
        onOpenChange={setComposeOpen}
      />
    </div>
  );
}
