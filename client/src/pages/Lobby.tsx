import { useEffect, useRef, useState } from 'react';
import AnimatedBackground from '@/components/AnimatedBackground';
import { useGame } from '../context/GameContext';
import { useNavigate } from 'react-router-dom';
import { WSEventPlayers, WSEventStart, WSEventChat } from '@/lib/constants';
import type { ChatMessage } from '@/types/types';

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Copy, Check, MessageCircle, X, Send } from "lucide-react";

interface Player {
  username: string;
  color: string;
}

const PLAYER_AVATARS = [
  "🦊", "🐸", "🦉", "🐙", "🦄", "🐲", "🦋", "🐺",
  "🦈", "🐢", "🦜", "🐼", "🦁", "🐨", "🦝", "🐯",
];

// Deterministic avatar per player
function getAvatar(username: string) {
  let hash = 0;
  for (let i = 0; i < username.length; i++) hash = (hash * 31 + username.charCodeAt(i)) | 0;
  return PLAYER_AVATARS[Math.abs(hash) % PLAYER_AVATARS.length];
}

// Deterministic float positions so they don't overlap too badly
function getFloatStyle(index: number, total: number) {
  const cols = Math.ceil(Math.sqrt(total));
  const row = Math.floor(index / cols);
  const col = index % cols;

  const cellW = 100 / cols;
  const cellH = 100 / Math.ceil(total / cols);

  // Center in cell with some jitter
  const jitterX = ((index * 17 + 7) % 11 - 5) * 1.5;
  const jitterY = ((index * 13 + 3) % 9 - 4) * 1.5;
  const left = cellW * col + cellW / 2 + jitterX;
  const top = cellH * row + cellH / 2 + jitterY;

  const duration = 6 + (index % 4) * 2;
  const delay = (index * 1.3) % 5;

  return { left: `${left}%`, top: `${top}%`, duration, delay };
}

export const Lobby: React.FC = () => {
  const navigate = useNavigate();
  const { ws, username, code, title, setTimeLeft } = useGame();
  const [players, setPlayers] = useState<Map<string, Player>>(new Map());
  const [creator, setCreator] = useState('');
  const [copied, setCopied] = useState(false);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [chatOpen, setChatOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);
  const [chatInput, setChatInput] = useState('');
  const chatEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!ws) return;
    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        if (message.Type === WSEventPlayers) {
          setPlayers((prev) => {
            const updated = new Map(prev);
            for (const [username, playerData] of Object.entries(message.Players)) {
              const player = playerData as { username?: string; color?: string };
              updated.set(username, {
                username: player.username ?? username,
                color: player.color ?? '#888888',
              });
            }
            return updated;
          });
          if (typeof message.Creator === 'string') {
            setCreator(message.Creator);
          }
        } else if (message.Type === WSEventChat) {
          const chat = message.Chat as ChatMessage;
          setChatMessages((prev) => [...prev, chat]);
          setChatOpen((open) => {
            if (!open) setUnreadCount((n) => n + 1);
            return open;
          });
        } else if (message.Type === WSEventStart) {
          ws.onmessage = null
          setTimeLeft(0);
          navigate('/game')
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };

    return () => { ws.onmessage = null; }
  }, [ws]);

  useEffect(() => {
    if (chatOpen) chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chatMessages, chatOpen]);

  const copyCode = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  const toggleChat = () => {
    setChatOpen((open) => !open);
    setUnreadCount(0);
  };

  const sendChat = () => {
    const text = chatInput.trim();
    if (!text || ws?.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({ username, code, Message: text }));
    setChatInput('');
  };

  const isHost = !!username && username === creator;

  const startGame = () => {
    if (ws?.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({ username, code, StartGame: true }));
  };

  return (
    
    <div className="relative min-h-screen flex flex-col items-center overflow-hidden">
      <AnimatedBackground />
      {/* Top bar */}
      <header className="relative z-10 w-full max-w-2xl mx-auto px-6 pt-8 animate-fade-up">
        <div className="card-glass rounded-2xl px-6 py-5 flex flex-col gap-4">
          {/* Game name */}
          <h1 className="font-display text-2xl font-bold title-gradient text-center leading-tight">
            {title}
          </h1>

          {/* Code + countdown row */}
          <div className="flex items-center justify-between gap-4">
            {/* Game code */}
            <button
              onClick={copyCode}
              className="flex items-center gap-2 bg-input/60 border border-border rounded-xl px-4 py-2 transition-all hover:border-primary/40 group"
            >
              <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Code</span>
              <span className="font-display text-lg font-bold tracking-[0.2em] text-foreground">
                {code}
              </span>
              {copied ? (
                <Check className="w-4 h-4 text-green-400" />
              ) : (
                <Copy className="w-4 h-4 text-muted-foreground group-hover:text-foreground transition-colors" />
              )}
            </button>

            {/* Start control */}
            {isHost ? (
              <button
                onClick={startGame}
                disabled={players.size === 0}
                className="btn-primary text-sm px-4 py-2 disabled:opacity-40 disabled:cursor-not-allowed"
              >
                Start Game
              </button>
            ) : (
              <span className="text-xs text-muted-foreground text-right max-w-[140px]">
                Waiting for host to start…
              </span>
            )}
          </div>

          {/* Player count */}
          <p className="text-sm text-muted-foreground text-center">
            <span className="text-foreground font-semibold">{players.size}</span> player{players.size !== 1 ? "s" : ""} joined
          </p>
        </div>
      </header>

      {/* Floating players area */}
      <div className="relative z-10 flex-1 w-full max-w-3xl mx-auto mt-6 mb-8 px-4">
        <div className="relative w-full h-[55vh]">
          {Array.from(players.values()).map((player, i) => {
            const pos = getFloatStyle(i, players.size);
            return (
              <div
                key={player.username}
                className="absolute -translate-x-1/2 -translate-y-1/2 flex flex-col items-center gap-1.5"
                style={{
                  left: pos.left,
                  top: pos.top,
                  animation: `lobby-float ${pos.duration}s ease-in-out ${pos.delay}s infinite`,
                }}
              >
                <Avatar
                  className="h-14 w-14 ring-2 ring-offset-2 ring-offset-background transition-transform hover:scale-110"
                  style={{
                    boxShadow: `0 0 18px hsl(${player.color} / 0.45)`,
                    ringColor: `hsl(${player.color})`,
                    // @ts-ignore ring-color via style
                    "--tw-ring-color": `hsl(${player.color})`,
                  } as React.CSSProperties}
                >
                  <AvatarFallback
                    className="text-2xl"
                    style={{ background: `hsl(${player.color} / 0.2)` }}
                  >
                    {getAvatar(player.username)}
                  </AvatarFallback>
                </Avatar>
                <span
                  className="text-xs font-semibold px-2 py-0.5 rounded-full backdrop-blur-sm"
                  style={{
                    color: `hsl(${player.color})`,
                    background: `hsl(${player.color} / 0.12)`,
                    border: `1px solid hsl(${player.color} / 0.25)`,
                  }}
                >
                  {player.username}
                </span>
              </div>
            );
          })}
        </div>
      </div>

      {/* Chat toggle button */}
      <button
        onClick={toggleChat}
        className="fixed bottom-6 right-6 z-30 flex items-center justify-center h-14 w-14 rounded-full card-glass border border-border shadow-lg hover:border-primary/40 transition-all"
      >
        {chatOpen ? (
          <X className="w-6 h-6 text-foreground" />
        ) : (
          <MessageCircle className="w-6 h-6 text-foreground" />
        )}
        {!chatOpen && unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 min-w-[20px] h-5 px-1 rounded-full bg-primary text-primary-foreground text-xs font-bold flex items-center justify-center">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        )}
      </button>

      {/* Chat panel */}
      {chatOpen && (
        <div className="fixed bottom-24 right-6 z-30 w-[calc(100vw-3rem)] max-w-sm h-[60vh] card-glass rounded-2xl border border-border flex flex-col overflow-hidden animate-fade-up">
          <div className="px-4 py-3 border-b border-border">
            <h2 className="font-display text-sm font-bold text-foreground">Lobby chat</h2>
          </div>
          <div className="flex-1 overflow-y-auto px-4 py-3 flex flex-col gap-2">
            {chatMessages.length === 0 ? (
              <p className="text-xs text-muted-foreground text-center mt-4">No messages yet — say hi!</p>
            ) : (
              chatMessages.map((msg, i) => (
                <div key={i} className="text-sm leading-snug break-words">
                  <span className="font-semibold" style={{ color: `hsl(${msg.color})` }}>
                    {msg.username}
                  </span>
                  <span className="text-foreground">: {msg.text}</span>
                </div>
              ))
            )}
            <div ref={chatEndRef} />
          </div>
          <div className="p-3 border-t border-border flex items-center gap-2">
            <input
              value={chatInput}
              onChange={(e) => setChatInput(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') sendChat(); }}
              placeholder="Type a message..."
              maxLength={280}
              className="flex-1 bg-input/60 border border-border rounded-xl px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:border-primary/40"
            />
            <button
              onClick={sendChat}
              disabled={!chatInput.trim()}
              className="flex items-center justify-center h-9 w-9 rounded-xl bg-primary/90 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-primary transition-colors shrink-0"
            >
              <Send className="w-4 h-4 text-primary-foreground" />
            </button>
          </div>
        </div>
      )}

      {/* Float keyframes */}
      <style>{`
        @keyframes lobby-float {
          0%, 100% { transform: translate(-50%, -50%) translateY(0px) translateX(0px); }
          25% { transform: translate(-50%, -50%) translateY(-10px) translateX(5px); }
          50% { transform: translate(-50%, -50%) translateY(4px) translateX(-7px); }
          75% { transform: translate(-50%, -50%) translateY(-6px) translateX(3px); }
        }
      `}</style>
    </div>
  );

};