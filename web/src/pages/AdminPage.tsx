import { useState } from "react";
import Sidebar from "../components/sidebar";
import AdminUsers from "../components/admin-users";
import AdminMailboxes from "../components/admin-mailboxes";
import ComposeDialog from "../components/compose-dialog";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "../components/ui/tabs";

export default function AdminPage() {
  const [composeOpen, setComposeOpen] = useState(false);

  return (
    <div className="flex h-screen">
      <Sidebar onCompose={() => setComposeOpen(true)} />
      <div className="flex-1 p-6 overflow-y-auto">
        <h1 className="text-2xl font-bold mb-6">Admin</h1>
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
