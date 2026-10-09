import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../lib/api";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Mail, ShieldCheck, Zap, Users, AlertCircle, Loader2 } from "lucide-react";

const features = [
  { icon: Zap, title: "Fast & focused", text: "A clean inbox that stays out of your way." },
  { icon: Users, title: "Shared mailboxes", text: "Give your team access to the right addresses." },
  { icon: ShieldCheck, title: "Secure by default", text: "Admin-controlled access to every mailbox." },
];

export default function LoginPage() {
  const navigate = useNavigate();
  const [setup, setSetup] = useState(false);
  const [checked, setChecked] = useState(false);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    api<{ needsSetup: boolean }>("/auth/setup-status")
      .then((s) => setSetup(s.needsSetup))
      .catch(() => {})
      .finally(() => setChecked(true));
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      if (setup) {
        await api("/auth/setup", {
          method: "POST",
          body: JSON.stringify({ name, email, password }),
        });
      }
      await api("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });
      navigate("/dashboard");
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  if (!checked) return null;

  return (
    <div className="min-h-screen lg:grid lg:grid-cols-2">
      {/* Brand panel */}
      <div className="relative hidden overflow-hidden bg-sidebar p-12 text-white lg:flex lg:flex-col lg:justify-between">
        <div className="pointer-events-none absolute -left-24 -top-24 h-96 w-96 rounded-full bg-indigo-500/30 blur-3xl" />
        <div className="pointer-events-none absolute -bottom-32 -right-16 h-[28rem] w-[28rem] rounded-full bg-violet-500/25 blur-3xl" />
        <div className="relative flex items-center gap-2.5 text-lg font-semibold tracking-tight">
          <span className="grid h-9 w-9 place-items-center rounded-xl bg-gradient-to-br from-indigo-400 to-violet-500 shadow-lg shadow-indigo-900/40">
            <Mail className="h-5 w-5" />
          </span>
          Email App
        </div>

        <div className="relative max-w-md">
          <h2 className="text-4xl font-semibold leading-tight tracking-tight">
            All your team email, in one calm place.
          </h2>
          <p className="mt-4 text-base text-sidebar-muted">
            Send, receive and manage mailboxes for your domain without the clutter.
          </p>
          <ul className="mt-10 space-y-6">
            {features.map(({ icon: Icon, title, text }) => (
              <li key={title} className="flex gap-4">
                <span className="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-white/10 ring-1 ring-white/10">
                  <Icon className="h-5 w-5 text-indigo-300" />
                </span>
                <div>
                  <p className="font-medium">{title}</p>
                  <p className="text-sm text-sidebar-muted">{text}</p>
                </div>
              </li>
            ))}
          </ul>
        </div>

        <p className="relative text-xs text-sidebar-muted">© {new Date().getFullYear()} Email App</p>
      </div>

      {/* Form panel */}
      <div className="flex min-h-screen items-center justify-center bg-background px-6 py-12 lg:min-h-0">
        <div className="w-full max-w-sm">
          <div className="mb-8 flex items-center gap-2.5 text-lg font-semibold tracking-tight lg:hidden">
            <span className="grid h-9 w-9 place-items-center rounded-xl bg-gradient-to-br from-indigo-500 to-violet-500 text-white">
              <Mail className="h-5 w-5" />
            </span>
            Email App
          </div>

          <h1 className="text-2xl font-semibold tracking-tight">
            {setup ? "Create admin account" : "Welcome back"}
          </h1>
          <p className="mt-1.5 text-sm text-muted-foreground">
            {setup
              ? "Set up the first administrator to get started."
              : "Sign in to access your mailboxes."}
          </p>

          <form onSubmit={handleSubmit} className="mt-8 space-y-4">
            {error && (
              <div className="flex items-start gap-2 rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">
                <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                {error}
              </div>
            )}
            {setup && (
              <div className="space-y-1.5">
                <label htmlFor="name" className="text-sm font-medium">Name</label>
                <Input id="name" placeholder="Your name" value={name} onChange={(e) => setName(e.target.value)} required />
              </div>
            )}
            <div className="space-y-1.5">
              <label htmlFor="email" className="text-sm font-medium">Email</label>
              <Input id="email" type="email" placeholder="you@company.com" value={email} onChange={(e) => setEmail(e.target.value)} required />
            </div>
            <div className="space-y-1.5">
              <label htmlFor="password" className="text-sm font-medium">Password</label>
              <Input id="password" type="password" placeholder="••••••••" value={password} onChange={(e) => setPassword(e.target.value)} required />
            </div>
            <Button type="submit" className="w-full h-11" disabled={loading}>
              {loading && <Loader2 className="animate-spin" />}
              {loading ? (setup ? "Creating..." : "Signing in...") : setup ? "Create admin" : "Sign in"}
            </Button>
          </form>
        </div>
      </div>
    </div>
  );
}
