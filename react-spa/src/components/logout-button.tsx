import { useNavigate } from "react-router-dom";
import { api } from "../lib/api";
import { Button } from "./ui/button";
import { LogOut } from "lucide-react";

export default function LogoutButton() {
  const navigate = useNavigate();

  async function handleLogout() {
    await api("/auth/logout", { method: "POST" });
    navigate("/login");
  }

  return (
    <Button variant="ghost" size="sm" onClick={handleLogout} className="w-full justify-start gap-2">
      <LogOut className="h-4 w-4" />
      Logout
    </Button>
  );
}
