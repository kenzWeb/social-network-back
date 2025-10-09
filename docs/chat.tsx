'use client';
import * as React from 'react';
import { useQueryClient } from '@tanstack/react-query';

type WsEvent =
  | { type: 'message'; payload: { conversationId: string; /* ...msg */ } }
  | { type: 'message.updated'; payload: { conversationId: string; id: string; /* patch */ } }
  | { type: 'typing'; payload: { conversationId: string; userId: string; isTyping: boolean } }
  | { type: 'presence'; payload: { userId: string; online: boolean } };

export function useReactQuerySubscription(opts: {
  token: string;
  conversationId?: string; // укажите, если бэк требует
  baseUrl?: string;        // опционально
}) {
  const { token, conversationId, baseUrl } = opts;
  const qc = useQueryClient();

  React.useEffect(() => {
    if (!token) return;

    const origin = baseUrl ?? window.location.origin;
    const wsScheme = origin.startsWith('https') ? 'wss' : 'ws';
    const qs = new URLSearchParams(
      conversationId ? { token, conversationId } : { token }
    ).toString();
    const url = origin.replace(/^http(s?):\/\//, `${wsScheme}://`) + `/ws/chat?` + qs;

    const ws = new WebSocket(url);

    ws.onmessage = (e) => {
      let data: WsEvent | undefined;
      try {
        data = JSON.parse(e.data);
      } catch { return; }
      if (!data) return;

      switch (data.type) {
        case 'message': {
          const cid = data.payload.conversationId;
          // Быстрый UX: сразу дописываем в кэш сообщения
          qc.setQueryData<any[]>(['chat','messages', cid], (old) =>
            old ? [...old, data!.payload] : [data!.payload]
          );
          // А также можно инвалидировать список диалогов (непрочитанные/последнее сообщение)
          qc.invalidateQueries({ queryKey: ['chat','conversations'], exact: false });
          break;
        }
        case 'message.updated': {
          const cid = data.payload.conversationId;
          qc.setQueryData<any[]>(['chat','messages', cid], (old) =>
            Array.isArray(old)
              ? old.map((m) => (m.id === data!.payload.id ? { ...m, ...data!.payload } : m))
              : old
          );
          break;
        }
        case 'typing': {
          const cid = data.payload.conversationId;
          qc.setQueryData(['chat','typing', cid], data.payload);
          break;
        }
        case 'presence': {
          qc.setQueryData(['chat','presence'], (old: Record<string, boolean> = {}) => ({
            ...old,
            [data!.payload.userId]: !!data!.payload.online,
          }));
          break;
        }
      }
    };

    return () => ws.close();
  }, [token, conversationId, baseUrl, qc]);
}