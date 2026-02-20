import { useEffect, useState } from 'react';
import AnimatedBackground from '@/components/AnimatedBackground';
import { useGame } from '../context/GameContext';
import { useNavigate } from 'react-router-dom';

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Progress } from "@/components/ui/progress";
import { Copy, Check } from "lucide-react";

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
  const { ws, code} = useGame();
  const [timeLeft, setTimeLeft] = useState<number | null>(null);
  const [players, setPlayers] = useState<Map<string, Player>>(new Map());
  const [copied, setCopied] = useState(false);

  var initialTime = 0

  useEffect(() => {
    if (!ws) return;
    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        if (message.Type === 'Time') {
          if (initialTime == 0) {
            initialTime = message.TimeLeft
          }
          setTimeLeft(message.TimeLeft);
        } else if (message.Type === 'Players') {
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
        } else if (message.Type === 'Start') {
          // Game hasn't been developed yet — return to home screen
          navigate('/game')
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };

    return () => { ws.onmessage = null; }
  }, [ws]);

  const copyCode = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div className="relative min-h-screen flex flex-col items-center overflow-hidden">
      <AnimatedBackground />

      {/* Top bar */}
      <header className="relative z-10 w-full max-w-2xl mx-auto px-6 pt-8 animate-fade-up">
        <div className="card-glass rounded-2xl px-6 py-5 flex flex-col gap-4">
          {/* Game name */}
          <h1 className="font-display text-2xl font-bold title-gradient text-center leading-tight">
            {"GAME_NAME"}
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

            {/* Countdown */}
            <div className="flex flex-col items-end gap-1 min-w-[120px]">
              <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Starting in
              </span>
              <div className="flex items-baseline gap-1">
                <span className="font-display text-3xl font-bold text-foreground tabular-nums" style={{color: (timeLeft !== null && timeLeft <= 3) ? "red" : ""}}>
                  {timeLeft}
                </span>
                <span className="text-sm text-muted-foreground">s</span>
              </div>
              
              <Progress
                value={(timeLeft !== null && initialTime != 0) ? 100 - ((timeLeft / initialTime) * 100) : 0}
                dir="rtl"
                className="h-1.5 w-full bg-muted/50"
              />
            </div>
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