import { useEffect, useRef, useCallback } from 'react';
import { useChatStore } from '../../store/chatStore';
import { useWebSocket } from '../../hooks/useWebSocket';
import { chatApi, type Message } from '../../api/chat';
import { MessageBubble } from './MessageBubble';
import { ChatInput } from './ChatInput';

interface ChatTimelineProps {
  projectId: string;
}

export function ChatTimeline({ projectId }: ChatTimelineProps) {
  const { messages, fetchHistory, addMessage, setConnected } = useChatStore();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    fetchHistory(projectId);
    return () => useChatStore.getState().clearMessages();
  }, [projectId, fetchHistory]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const tokens = localStorage.getItem('tokens');
  const accessToken = tokens ? JSON.parse(tokens).access_token : '';
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = accessToken
    ? `${wsProtocol}//${window.location.host}/api/v1/projects/${projectId}/chat/ws?token=${accessToken}`
    : '';

  const handleWsMessage = useCallback((data: string) => {
    try {
      const msg: Message = JSON.parse(data);
      addMessage(msg);
    } catch {
      // ignore non-JSON messages
    }
  }, [addMessage]);

  const { send } = useWebSocket({
    url: wsUrl,
    onMessage: handleWsMessage,
    onOpen: () => setConnected(true),
    onClose: () => setConnected(false),
    enabled: !!accessToken,
  });

  const handleSend = async (content: string) => {
    // Send via REST (also broadcasts via WS hub)
    await chatApi.sendMessage(projectId, content);
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 overflow-y-auto py-2">
        {messages.length === 0 ? (
          <div className="flex items-center justify-center h-full text-[var(--color-text-secondary)] text-sm">
            No messages yet. Start the conversation!
          </div>
        ) : (
          messages.map((msg) => <MessageBubble key={msg.id} message={msg} />)
        )}
        <div ref={messagesEndRef} />
      </div>
      <ChatInput onSend={handleSend} />
    </div>
  );
}
