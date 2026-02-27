import { User, Bot, Cpu, GitBranch, Wrench } from 'lucide-react';
import type { Message } from '../../api/chat';

const senderIcons = {
  human: User,
  agent: Bot,
  system: Cpu,
  build: Wrench,
  repo: GitBranch,
};

const senderColors = {
  human: 'bg-[var(--color-primary)]/20 text-[var(--color-primary)]',
  agent: 'bg-purple-500/20 text-purple-400',
  system: 'bg-gray-500/20 text-gray-400',
  build: 'bg-yellow-500/20 text-yellow-400',
  repo: 'bg-green-500/20 text-green-400',
};

export function MessageBubble({ message }: { message: Message }) {
  const Icon = senderIcons[message.sender_type] || Cpu;
  const colorClass = senderColors[message.sender_type] || senderColors.system;
  const isHuman = message.sender_type === 'human';
  const time = new Date(message.created_at).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <div className={`flex gap-3 px-4 py-2 ${isHuman ? 'flex-row-reverse' : ''}`}>
      <div className={`shrink-0 w-8 h-8 rounded-full flex items-center justify-center ${colorClass}`}>
        <Icon size={16} />
      </div>
      <div className={`max-w-[80%] ${isHuman ? 'items-end' : 'items-start'} flex flex-col gap-1`}>
        <div className="flex items-center gap-2 text-xs text-[var(--color-text-secondary)]">
          <span className="font-medium">{message.sender_name || message.sender_type}</span>
          <span>{time}</span>
        </div>
        <div
          className={`rounded-xl px-3 py-2 text-sm ${
            isHuman
              ? 'bg-[var(--color-primary)] text-white rounded-tr-sm'
              : 'bg-[var(--color-bg-tertiary)] rounded-tl-sm'
          }`}
        >
          {message.message_type === 'code' || message.message_type === 'agent_output' ? (
            <pre className="whitespace-pre-wrap font-mono text-xs">{message.content}</pre>
          ) : (
            <p className="whitespace-pre-wrap">{message.content}</p>
          )}
        </div>
      </div>
    </div>
  );
}
