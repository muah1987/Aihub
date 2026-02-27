import { useEffect, useRef, useCallback } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';

interface UseTerminalOptions {
  wsUrl: string;
  enabled?: boolean;
}

export function useTerminal({ wsUrl, enabled = true }: UseTerminalOptions) {
  const terminalRef = useRef<Terminal | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);

  const attach = useCallback((container: HTMLDivElement) => {
    containerRef.current = container;

    if (terminalRef.current) {
      terminalRef.current.dispose();
    }

    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: '"Fira Code", "JetBrains Mono", monospace',
      theme: {
        background: '#0f172a',
        foreground: '#f8fafc',
        cursor: '#6366f1',
        selectionBackground: '#334155',
      },
      rows: 24,
      cols: 80,
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.loadAddon(new WebLinksAddon());

    term.open(container);
    fitAddon.fit();

    terminalRef.current = term;
    fitAddonRef.current = fitAddon;

    // Connect WebSocket
    if (enabled && wsUrl) {
      const ws = new WebSocket(wsUrl);
      ws.binaryType = 'arraybuffer';
      wsRef.current = ws;

      ws.onopen = () => {
        term.writeln('\x1b[32mConnected to sandbox\x1b[0m');
      };

      ws.onmessage = (event) => {
        const data = event.data instanceof ArrayBuffer
          ? new Uint8Array(event.data)
          : event.data;
        term.write(data);
      };

      ws.onclose = () => {
        term.writeln('\x1b[31mDisconnected from sandbox\x1b[0m');
      };

      term.onData((data) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(data);
        }
      });
    }

    // Handle resize
    const resizeObserver = new ResizeObserver(() => {
      fitAddon.fit();
    });
    resizeObserver.observe(container);

    return () => {
      resizeObserver.disconnect();
      wsRef.current?.close();
      term.dispose();
    };
  }, [wsUrl, enabled]);

  useEffect(() => {
    return () => {
      wsRef.current?.close();
      terminalRef.current?.dispose();
    };
  }, []);

  return { attach };
}
