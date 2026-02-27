import { useAuth } from '../../hooks/useAuth';
import { LogOut, Settings, Cpu } from 'lucide-react';

export function Header() {
  const { user, logout } = useAuth();

  return (
    <header className="h-14 bg-[var(--color-bg-secondary)] border-b border-[var(--color-border)] flex items-center justify-between px-4 shrink-0">
      <div className="flex items-center gap-2">
        <Cpu size={24} className="text-[var(--color-primary)]" />
        <span className="font-bold text-lg">Aihub</span>
      </div>

      {user && (
        <div className="flex items-center gap-3">
          <span className="text-sm text-[var(--color-text-secondary)] hidden sm:block">
            {user.display_name}
          </span>
          <button
            className="p-2 hover:bg-[var(--color-bg-tertiary)] rounded-lg"
            title="Settings"
          >
            <Settings size={18} />
          </button>
          <button
            onClick={logout}
            className="p-2 hover:bg-[var(--color-bg-tertiary)] rounded-lg text-[var(--color-text-secondary)]"
            title="Logout"
          >
            <LogOut size={18} />
          </button>
        </div>
      )}
    </header>
  );
}
