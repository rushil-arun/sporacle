import { useState, useEffect } from "react";
import AnimatedBackground from "@/components/AnimatedBackground";
import { Progress } from "@/components/ui/progress";
import { useGame } from '../context/GameContext';
import { useNavigate } from "react-router-dom";

// ── Mock config ────────────────────────────────────────────────────────────────
const CATEGORY = "Countries in Europe";
const TOTAL_SECONDS = 120;
const TOTAL_ITEMS = 44; // total answers possible

// Map<item, { username, color }>
type PlayerMeta = { username: string; color: string };
type Request = { username: string, code: string, Item: string }
const INITIAL_BOARD = new Map<string, PlayerMeta>([
]);

// ── Helpers ────────────────────────────────────────────────────────────────────
// Build display grid: filled cells first, then blank placeholders
function buildGrid(board: Map<string, PlayerMeta>, total: number) {
  const filled = Array.from(board.entries()).map(([item, player]) => ({ item, player }));
  const blanks = Array.from({ length: Math.max(0, total - filled.length) }, () => ({
    item: null as string | null,
    player: null as PlayerMeta | null,
  }));
  return [...filled, ...blanks];
}

// Optimal cols/rows to fill a grid with `total` cells
function gridDimensions(total: number) {
  const cols = Math.ceil(Math.sqrt(total));
  const rows = Math.ceil(total / cols);
  return { cols, rows };
}

// ── Component ──────────────────────────────────────────────────────────────────
export const Game: React.FC = () => {
  const navigate = useNavigate();
  const [timeLeft, setTimeLeft] = useState(TOTAL_SECONDS);
  const [board, setBoard] = useState<Map<string, PlayerMeta>>(INITIAL_BOARD);
  const { username, ws, code } = useGame();
  const [inputValue, setInputValue] = useState("");

  const handleSubmit = (item: string) => {
    const request = { username : username, code: code, Item: item}
    if (ws?.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(request));
    }
    setInputValue("")
  }

  useEffect(() => {
    if (timeLeft <= 0) return;
    const id = setInterval(() => setTimeLeft((t) => Math.max(0, t - 1)), 1000);
    return () => clearInterval(id);
  }, [timeLeft]);

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
        } else if (message.Type === 'Board') {
          setBoard((prev) => {
            const updated = new Map(prev);
            console.log(Object.entries(message.State))
            for (const [item, playerData] of Object.entries(message.State)) {
              const player = playerData == null ? { username : "", color: "#888888"} : playerData as { username?: string; color?: string };
              updated.set(item, {
                username: player.username ?? "",
                color: player.color ?? '#888888',
              });
            }
            return updated;
          });
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };

    return () => { ws.onmessage = null; }
  }, [ws]);

  const grid = buildGrid(board, TOTAL_ITEMS);
  const { cols, rows } = gridDimensions(TOTAL_ITEMS);
  const minutes = Math.floor(timeLeft / 60);
  const seconds = timeLeft % 60;
  const timeDisplay = `${minutes}:${seconds.toString().padStart(2, "0")}`;

  return (
    <div className="relative h-screen flex flex-col overflow-hidden">
      <AnimatedBackground />

      {/* ── Header ── */}
      <header className="relative z-10 w-full max-w-5xl mx-auto px-6 pt-6 animate-fade-up">
        <div className="card-glass rounded-2xl px-6 py-4 flex items-center justify-between gap-6">
          {/* Category */}
          <div className="flex flex-col gap-0.5 min-w-0">
            <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
              Category
            </span>
            <h1 className="font-display text-xl font-bold title-gradient leading-tight truncate">
              {CATEGORY}
            </h1>
          </div>

          {/* Timer */}
          <div className="flex flex-col items-end gap-2 shrink-0 min-w-[160px]">
            <div className="flex items-center gap-2">
              <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Time left
              </span>
              <span
                className="font-display text-4xl font-bold tabular-nums leading-none"
                style={{
                  color: timeLeft <= 10 ? "hsl(350 75% 62%)" : "hsl(var(--foreground))",
                  transition: "color 0.3s",
                }}
              >
                {timeDisplay}
              </span>
            </div>
            <Progress
              value={100 - ((timeLeft / 10) * 100)}
              className="h-2.5 w-full bg-muted/50"
            />
          </div>
        </div>
      </header>

      {/* ── Score bar ── */}
      <div className="relative z-10 w-full max-w-5xl mx-auto px-6 pt-3 animate-fade-up">
        <p className="text-sm text-muted-foreground text-center">
          <span className="text-foreground font-semibold">{board.size}</span> of{" "}
          <span className="text-foreground font-semibold">{TOTAL_ITEMS}</span> found
        </p>
      </div>

      {/* ── Grid ── */}
      <main className="relative z-10 flex-1 min-h-0 w-full max-w-5xl mx-auto px-6 pt-3 pb-3">
        <div
          className="grid gap-2 h-full w-full"
          style={{
            gridTemplateColumns: `repeat(${cols}, 1fr)`,
            gridTemplateRows: `repeat(${rows}, 1fr)`,
          }}
        >
          {grid.map((cell, i) =>
            cell.item && cell.player && cell.player?.username != "" ? (
              <FilledCell key={cell.item} item={cell.item} player={cell.player} />
            ) : (
              <div
                key={`blank-${i}`}
                className="rounded-xl border border-border/50 bg-muted/20 backdrop-blur-sm"
                style={{
                    border: `2px solid hsl(155 45% 85% / 0.4)`
                }}
              />
            )
          )}
        </div>
      </main>

      {/* ── Input bar ── */}
      <div className="relative z-10 w-full max-w-5xl mx-auto px-6 pb-6 animate-fade-up">
        <div className="card-glass rounded-2xl px-4 py-3 flex items-center gap-3">
          <input
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            placeholder="Type an answer…"
            onKeyDown={(e) => {
                if (e.key === "Enter") {
                    handleSubmit(inputValue);
                }
            }}
            autoFocus
            className="input-sporacle bg-transparent border-none shadow-none flex-1 text-base"
            style={{ boxShadow: "none" }}
          />
          <span className="text-xs text-muted-foreground shrink-0">Press Enter to submit</span>
        </div>
      </div>
    </div>
  );
};

// ── Filled cell sub-component ──────────────────────────────────────────────────
function FilledCell({ item, player }: { item: string; player: PlayerMeta }) {
  return (
    <div
      className="relative rounded-xl overflow-hidden flex flex-col items-center justify-center gap-1 p-2 text-center animate-scale-in"
      style={{
        background: `hsl(${player.color} / 0.18)`,
        border: `3px solid hsl(${player.color} / 0.4)`,
        boxShadow: `0 0 16px hsl(${player.color} / 0.2)`,
      }}
    >
      {/* Subtle color wash */}
      <div
        className="absolute inset-0 pointer-events-none"
        style={{
          background: `radial-gradient(ellipse at 50% 0%, hsl(${player.color} / 0.15) 0%, transparent 70%)`,
        }}
      />
      <span className="relative font-display font-semibold text-sm leading-tight" style={{ color: `hsl(${player.color})` }}>
        {item}
      </span>
      <span
        className="relative text-[10px] font-medium px-1.5 py-0.5 rounded-full"
        style={{
          background: `hsl(${player.color} / 0.2)`,
          color: `hsl(${player.color})`,
          border: `1px solid hsl(${player.color} / 0.3)`,
        }}
      >
        {player.username}
      </span>
    </div>
  );
}