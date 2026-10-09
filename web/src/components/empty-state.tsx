import { Mail } from "lucide-react";

export default function EmptyState({
  icon,
  title,
  description,
}: {
  icon?: React.ReactNode;
  title: string;
  description?: string;
}) {
  return (
    <div className="flex h-full flex-col items-center justify-center p-6 text-center">
      <div className="mb-5 grid h-20 w-20 place-items-center rounded-2xl bg-gradient-to-br from-indigo-50 to-violet-100 text-primary ring-1 ring-primary/10 [&_svg]:h-9 [&_svg]:w-9">
        {icon || <Mail />}
      </div>
      <p className="text-base font-semibold text-foreground">{title}</p>
      {description && <p className="mt-1 max-w-xs text-sm text-muted-foreground">{description}</p>}
    </div>
  );
}
