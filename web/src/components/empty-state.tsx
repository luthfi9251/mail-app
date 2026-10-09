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
    <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
      <div className="mb-4 opacity-20">
        {icon || <Mail className="h-16 w-16" />}
      </div>
      <p className="text-lg font-medium">{title}</p>
      {description && <p className="text-sm mt-1">{description}</p>}
    </div>
  );
}
