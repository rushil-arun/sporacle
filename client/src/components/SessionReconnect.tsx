import { useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useGame } from '@/context/GameContext';
import { WSHandshakeError, WSHandshakeSuccess } from '@/lib/constants';
import { clearSession, loadSession } from '@/lib/session';

const RETRY_DELAY_MS = 1000;
const MAX_RETRIES = 15;

// Keeps the player connected: restores the session after a page refresh and
// re-opens the WebSocket if it drops unexpectedly. Renders nothing.
export function SessionReconnect() {
  const navigate = useNavigate();
  const { ws, setWs, setUsername, setCode, setWsUrl, setTitle } = useGame();
  const retries = useRef(0);
  const retryTimer = useRef<number | undefined>(undefined);

  const connect = () => {
    const session = loadSession();
    if (!session) return;
    const sep = session.wsUrl.includes('?') ? '&' : '?';
    const socket = new WebSocket(`${session.wsUrl}${sep}token=${encodeURIComponent(session.token)}`);

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === WSHandshakeSuccess) {
          retries.current = 0;
          setUsername(session.username);
          setCode(session.code);
          setWsUrl(session.wsUrl);
          setTitle(data.message ?? session.title);
          setWs(socket);
          navigate(data.started ? '/game' : '/lobby');
        } else if (data.type === WSHandshakeError) {
          // The game is gone or the token was rejected; nothing to resume.
          clearSession();
          socket.onclose = null;
          navigate('/');
        }
      } catch {
        // ignore malformed handshake
      }
    };
    socket.onclose = scheduleRetry;
  };

  function scheduleRetry() {
    if (!loadSession() || retries.current >= MAX_RETRIES) return;
    retries.current += 1;
    retryTimer.current = window.setTimeout(connect, RETRY_DELAY_MS);
  }

  // Resume after a refresh: a session is stored but there is no live socket.
  useEffect(() => {
    if (!ws && loadSession()) connect();
    return () => window.clearTimeout(retryTimer.current);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Reconnect if the live socket drops while the session is still active.
  useEffect(() => {
    if (!ws) return;
    ws.onclose = scheduleRetry;
    return () => {
      ws.onclose = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ws]);

  return null;
}
