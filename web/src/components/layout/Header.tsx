import { useState, useEffect } from 'react';
import { useAuth } from '../../hooks/useAuth';
import { LogOut, Settings, Cpu, Search } from 'lucide-react';
import { EmailVerificationBanner } from '../auth/EmailVerificationBanner';
import { NotificationBell } from '../notifications/NotificationBell';
import { GlobalSearch } from '../platform/GlobalSearch';

export function Header() {
  const { user, logout } = useAuth();
  const [searchOpen, setSearchOpen] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setSearchOpen((prev) => !prev);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  return (
    <>
      <header className="h-14 bg-[var(--color-bg-secondary)] border-b border-[var(--color-border)] flex items-center justify-between px-4 shrink-0">
        <div className="flex items-center gap-2">
          <Cpu size={24} className="text-[var(--color-primary)]" />
          <span className="font-bold text-lg">Aihub</span>
        </div>

        {user && (
          <div className="flex items-center gap-2">
            <button
              onClick={() => setSearchOpen(true)}
              className="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-[var(--color-bg-tertiary)] rounded-lg text-sm text-[var(--color-text-secondary)] hover:text-[var(--color-text)] transition-colors"
            >
              <Search size={14} />
              <span>Search</span>
              <kbd className="text-[10px] border border-[var(--color-border)] rounded px-1 py-0.5">&#8984;K</kbd>
            </button>
            <button
              onClick={() => setSearchOpen(true)}
              className="sm:hidden p-2 hover:bg-[var(--color-bg-tertiary)] rounded-lg"
              title="Search"
            >
              <Search size={18} />
            </button>
            <span className="text-sm text-[var(--color-text-secondary)] hidden sm:block">
              {user.display_name}
            </span>
            <NotificationBell />
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
      {user && !user.email_verified && <EmailVerificationBanner />}
      <GlobalSearch open={searchOpen} onClose={() => setSearchOpen(false)} />
    </>
  );
}
