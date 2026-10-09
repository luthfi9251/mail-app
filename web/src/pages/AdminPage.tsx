import { useState } from "react";
import Sidebar from "../components/sidebar";
import AdminUsers from "../components/admin-users";
import AdminMailboxes from "../components/admin-mailboxes";
import ComposeDialog from "../components/compose-dialog";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "../components/ui/tabs";
import { Button } from "../components/ui/button";
import { Menu, X } from "lucide-react";

export default function AdminPage() {
  const [composeOpen, setComposeOpen] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

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

      <div className="flex-1 p-6 overflow-y-auto min-w-0">
        <div className="flex items-center gap-2 mb-6">
          <Button variant="ghost" size="icon" className="lg:hidden -ml-2" onClick={() => setSidebarOpen(!sidebarOpen)}>
            {sidebarOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </Button>
          <h1 className="text-2xl font-bold">Admin</h1>
        </div>
        <Tabs defaultValue="users">
          <TabsList>
            <TabsTrigger value="users">Users</TabsTrigger>
            <TabsTrigger value="mailboxes">Mailboxes</TabsTrigger>
          </TabsList>
          <TabsContent value="users"><AdminUsers /></TabsContent>
          <TabsContent value="mailboxes"><AdminMailboxes /></TabsContent>
        </Tabs>
      </div>
      <ComposeDialog
        mailboxId=""
        open={composeOpen}
        onOpenChange={setComposeOpen}
      />
    </div>
  );
}
