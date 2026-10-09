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
  const [sidebarOpen, setSidebarOpen] = useState(false);

  if (!mailboxId) return null;

  return (
    <div className="flex h-screen overflow-hidden">
      <div className="hidden lg:flex">
        <Sidebar onCompose={() => setComposeOpen(true)} />
      </div>

      <div className={`fixed inset-0 z-40 lg:hidden ${sidebarOpen ? "block" : "hidden"}`}>
        <div className="absolute inset-0 bg-black/50" onClick={() => setSidebarOpen(false)} />
        <div className="relative w-64 h-full">
          <Sidebar
            onCompose={() => { setComposeOpen(true); setSidebarOpen(false); }}
            onClose={() => setSidebarOpen(false)}
          />
        </div>
      </div>

      <div className={`lg:flex ${emailId ? "hidden lg:flex" : "flex flex-1"}`}>
        <EmailList mailboxId={mailboxId} onToggleSidebar={() => setSidebarOpen(!sidebarOpen)} />
      </div>

      <div className={`flex-1 flex flex-col h-full min-w-0 overflow-hidden ${emailId ? "flex" : "hidden lg:flex"}`}>
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
