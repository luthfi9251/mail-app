import { useState } from "react";
import { useParams } from "react-router-dom";
import Sidebar from "../components/sidebar";
import MobileDrawer from "../components/mobile-drawer";
import EmailList from "../components/email-list";
import ThreadView from "../components/thread-view";
import ComposeDialog from "../components/compose-dialog";
import EmptyState from "../components/empty-state";

export default function InboxPage() {
  const { mailboxId, emailId } = useParams();
  const [composeOpen, setComposeOpen] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  if (!mailboxId) return null;

  return (
    <div className="flex h-dvh overflow-hidden">
      <div className="hidden lg:flex">
        <Sidebar onCompose={() => setComposeOpen(true)} />
      </div>

      <MobileDrawer open={sidebarOpen} onClose={() => setSidebarOpen(false)} onCompose={() => setComposeOpen(true)} />

      <div className={emailId ? "hidden lg:flex lg:shrink-0" : "flex flex-1 min-w-0 lg:flex-none lg:shrink-0"}>
        <EmailList mailboxId={mailboxId} onToggleSidebar={() => setSidebarOpen(!sidebarOpen)} />
      </div>

      <div className={`flex-1 flex flex-col h-full min-w-0 overflow-hidden bg-card ${emailId ? "flex" : "hidden lg:flex"}`}>
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
