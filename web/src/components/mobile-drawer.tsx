import Sidebar from "./sidebar";

export default function MobileDrawer({
  open,
  onClose,
  onCompose,
}: {
  open: boolean;
  onClose: () => void;
  onCompose: () => void;
}) {
  return (
    <div className={`fixed inset-0 z-40 lg:hidden ${open ? "block" : "hidden"}`}>
      <div className="absolute inset-0 bg-slate-950/50 backdrop-blur-sm" onClick={onClose} />
      <div className="drawer-in relative h-full w-64 shadow-2xl">
        <Sidebar onCompose={() => { onCompose(); onClose(); }} onClose={onClose} />
      </div>
    </div>
  );
}
