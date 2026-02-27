import { ReactNode } from 'react';

interface CardProps {
  children: ReactNode;
  className?: string;
  onClick?: () => void;
}

export function Card({ children, className = '', onClick }: CardProps) {
  return (
    <div
      className={`bg-[var(--color-bg-secondary)] border border-[var(--color-border)] rounded-xl p-4 ${onClick ? 'cursor-pointer hover:border-[var(--color-primary)] transition-colors' : ''} ${className}`}
      onClick={onClick}
    >
      {children}
    </div>
  );
}
