import { ReactNode } from 'react';
import { Header } from './Header';
import { Sidebar } from './Sidebar';
import { TabBar } from './TabBar';

interface AppShellProps {
  children: ReactNode;
  activeTab: string;
  onTabChange: (tab: string) => void;
  showTabs?: boolean;
}

export function AppShell({ children, activeTab, onTabChange, showTabs = false }: AppShellProps) {
  return (
    <div className="h-screen flex flex-col">
      <Header />
      <div className="flex flex-1 overflow-hidden">
        {showTabs && <Sidebar activeTab={activeTab} onTabChange={onTabChange} />}
        <main className="flex-1 overflow-y-auto">{children}</main>
      </div>
      {showTabs && <TabBar activeTab={activeTab} onTabChange={onTabChange} />}
    </div>
  );
}
