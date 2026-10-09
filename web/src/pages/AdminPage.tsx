import { useState } from "react";
import Sidebar from "../components/sidebar";
import MobileDrawer from "../components/mobile-drawer";
import AdminUsers from "../components/admin-users";
import AdminMailboxes from "../components/admin-mailboxes";
import ComposeDialog from "../components/compose-dialog";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "../components/ui/tabs";
import { Button } from "../components/ui/button";
import { Menu, Users, Inbox } from "lucide-react";

export default function AdminPage() {
  const [composeOpen, setComposeOpen] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className="flex h-dvh overflow-hidden">
      <div className="hidden lg:flex">
        <Sidebar onCompose={() => setComposeOpen(true)} />
      </div>

      <MobileDrawer open={sidebarOpen} onClose={() => setSidebarOpen(false)} onCompose={() => setComposeOpen(true)} />

      <div className="flex-1 min-w-0 overflow-y-auto scroll-thin">
        <div className="mx-auto max-w-5xl px-4 py-6 sm:px-8 sm:py-10">
          <div className="mb-8 flex items-start gap-3">
            <Button variant="ghost" size="icon" className="lg:hidden -ml-2" onClick={() => setSidebarOpen(true)}>
              <Menu className="h-5 w-5" />
            </Button>
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">Administration</h1>
              <p className="mt-1 text-sm text-muted-foreground">Manage users and the mailboxes they can access.</p>
            </div>
          </div>
          <Tabs defaultValue="users">
            <TabsList>
              <TabsTrigger value="users" className="gap-2"><Users className="h-4 w-4" />Users</TabsTrigger>
              <TabsTrigger value="mailboxes" className="gap-2"><Inbox className="h-4 w-4" />Mailboxes</TabsTrigger>
            </TabsList>
            <TabsContent value="users"><AdminUsers /></TabsContent>
            <TabsContent value="mailboxes"><AdminMailboxes /></TabsContent>
          </Tabs>
        </div>
      </div>
      <ComposeDialog mailboxId="" open={composeOpen} onOpenChange={setComposeOpen} />
    </div>
  );
}
