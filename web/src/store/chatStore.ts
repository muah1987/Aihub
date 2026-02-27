import { create } from 'zustand';
import { chatApi, type Message } from '../api/chat';

interface ChatState {
  messages: Message[];
  loading: boolean;
  connected: boolean;
  fetchHistory: (projectId: string) => Promise<void>;
  addMessage: (message: Message) => void;
  setConnected: (connected: boolean) => void;
  clearMessages: () => void;
}

export const useChatStore = create<ChatState>((set, get) => ({
  messages: [],
  loading: false,
  connected: false,

  fetchHistory: async (projectId) => {
    set({ loading: true });
    try {
      const { data } = await chatApi.getHistory(projectId);
      set({ messages: data.messages || [], loading: false });
    } catch {
      set({ loading: false });
    }
  },

  addMessage: (message) => {
    const exists = get().messages.some((m) => m.id === message.id);
    if (!exists) {
      set({ messages: [...get().messages, message] });
    }
  },

  setConnected: (connected) => set({ connected }),
  clearMessages: () => set({ messages: [] }),
}));
