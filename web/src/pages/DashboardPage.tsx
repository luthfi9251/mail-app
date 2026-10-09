import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api, type Mailbox } from "../lib/api";
import Sidebar from "../components/sidebar";
import EmptyState from "../components/empty-state";
import { Skeleton } from "../components/ui/skeleton";

export default function DashboardPage() {
  const navigate = useNavigate();
  const [mailboxes, setMailboxes] = useState<Mailbox[] | null>(null);

  useEffect(() => {
    api<Mailbox[]>("/mailboxes")
      .then((res) => {
        const list = res ?? [];
        setMailboxes(list);
        if (list.length > 0) {
          navigate(`/dashboard/${list[0].id}`, { replace: true });
        }
      })
      .catch(() => navigate("/login"));
  }, []);

  if (mailboxes === null) {
    return (
      <div className="flex h-screen">
        <div className="w-64 bg-sidebar p-4 space-y-3">
          <Skeleton className="h-10 w-full bg-sidebar-accent" />
          <Skeleton className="h-8 w-full bg-sidebar-accent" />
          <Skeleton className="h-8 w-full bg-sidebar-accent" />
        </div>
        <div className="flex-1 p-6 space-y-4">
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-20 w-full" />
          <Skeleton className="h-20 w-full" />
        </div>
      </div>
    );
  }

  if (mailboxes?.length === 0) {
    return (
      <div className="flex h-screen">
        <Sidebar onCompose={() => {}} />
        <div className="flex-1 flex items-center justify-center">
          <EmptyState
            title="No mailboxes"
            description="Contact your administrator to get access."
          />
        </div>
      </div>
    );
  }

  return null;
}
