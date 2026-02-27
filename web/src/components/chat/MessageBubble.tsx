import { User, Bot, Cpu, GitBranch, Wrench, Users, CheckCircle } from 'lucide-react';
import type { Message } from '../../api/chat';

const senderIcons: Record<string, typeof Cpu> = {
  human: User,
  agent: Bot,
  system: Cpu,
  build: Wrench,
  repo: GitBranch,
  team: Users,
};

const senderColors: Record<string, string> = {
  human: 'bg-[var(--color-primary)]/20 text-[var(--color-primary)]',
  agent: 'bg-purple-500/20 text-purple-400',
  system: 'bg-gray-500/20 text-gray-400',
  build: 'bg-yellow-500/20 text-yellow-400',
  repo: 'bg-green-500/20 text-green-400',
  team: 'bg-cyan-500/20 text-cyan-400',
};

const isCodeType = (type: string) =>
  type === 'code' || type === 'agent_output' || type === 'agent_task_complete';

const isTeamType = (type: string) =>
  type === 'team_synthesis' || type === 'team_invocation';

export function MessageBubble({ message }: { message: Message }) {
  const Icon = senderIcons[message.sender_type] || Cpu;
  const colorClass = senderColors[message.sender_type] || senderColors.system;
  const isHuman = message.sender_type === 'human';
  const time = new Date(message.created_at).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
  });

  const renderContent = () => {
    if (isCodeType(message.message_type)) {
      return <pre className="whitespace-pre-wrap font-mono text-xs">{message.content}</pre>;
    }

    if (message.message_type === 'team_synthesis') {
      return (
        <div>
          <div className="flex items-center gap-1.5 mb-1 text-xs font-medium text-cyan-400">
            <CheckCircle size={12} /> Team Synthesis
          </div>
          <p className="whitespace-pre-wrap">{message.content}</p>
        </div>
      );
    }

    if (isTeamType(message.message_type)) {
      return (
        <div>
          <div className="flex items-center gap-1.5 mb-1 text-xs font-medium text-cyan-400">
            <Users size={12} /> Team Task
          </div>
          <p className="whitespace-pre-wrap">{message.content}</p>
        </div>
      );
    }

    return <p className="whitespace-pre-wrap">{message.content}</p>;
  };

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
          {renderContent()}
        </div>
      </div>
    </div>
  );
}
