import { createContext, useCallback, useContext, useState } from "react";
import { AlertTriangle, Loader2 } from "lucide-react";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";
import { Button } from "./ui/button";

interface ConfirmOptions {
  title: string;
  description: React.ReactNode;
  confirmLabel?: string;
  destructive?: boolean;
  // Runs when the user confirms. The dialog stays open with a spinner until it
  // settles, and shows the error message if it throws.
  onConfirm: () => Promise<void> | void;
}

const ConfirmContext = createContext<(opts: ConfirmOptions) => void>(() => {});

export function useConfirm() {
  return useContext(ConfirmContext);
}

export function ConfirmProvider({ children }: { children: React.ReactNode }) {
  const [opts, setOpts] = useState<ConfirmOptions | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const confirm = useCallback((o: ConfirmOptions) => {
    setError("");
    setBusy(false);
    setOpts(o);
  }, []);

  async function handleConfirm() {
    if (!opts) return;
    setBusy(true);
    setError("");
    try {
      await opts.onConfirm();
      setOpts(null);
    } catch (err: any) {
      setError(err?.message ?? "Something went wrong");
    } finally {
      setBusy(false);
    }
  }

  const destructive = opts?.destructive ?? true;

  return (
    <ConfirmContext.Provider value={confirm}>
      {children}
      <Dialog open={opts !== null} onOpenChange={(open) => !open && !busy && setOpts(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader className="items-center sm:items-start">
            {destructive && (
              <span className="mb-1 grid h-10 w-10 place-items-center rounded-full bg-destructive/10 text-destructive">
                <AlertTriangle className="h-5 w-5" />
              </span>
            )}
            <DialogTitle>{opts?.title}</DialogTitle>
          </DialogHeader>
          <div className="text-center text-sm text-muted-foreground [overflow-wrap:anywhere] sm:text-left">{opts?.description}</div>
          {error && (
            <p className="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">{error}</p>
          )}
          <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <Button variant="outline" onClick={() => setOpts(null)} disabled={busy}>
              Cancel
            </Button>
            <Button variant={destructive ? "destructive" : "default"} onClick={handleConfirm} disabled={busy}>
              {busy && <Loader2 className="animate-spin" />}
              {opts?.confirmLabel ?? "Confirm"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </ConfirmContext.Provider>
  );
}
