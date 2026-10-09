import { useNavigate } from "react-router-dom";
import { api } from "../lib/api";
import { LogOut } from "lucide-react";

export default function LogoutButton() {
  const navigate = useNavigate();

  async function handleLogout() {
    await api("/auth/logout", { method: "POST" });
    navigate("/login");
  }

  return (
    <button
      onClick={handleLogout}
      className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-sidebar-foreground/80 transition-colors hover:bg-red-500/15 hover:text-red-300"
    >
      <LogOut className="h-4 w-4 text-sidebar-muted" />
      Logout
    </button>
  );
}
