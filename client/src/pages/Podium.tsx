import { useNavigate } from "react-router-dom";
import AnimatedBackground from "@/components/AnimatedBackground";
import { Trophy } from "lucide-react";
import { useGame } from '../context/GameContext';
import type { LeaderboardEntry } from "@/types/types";



const Podium = () => {
  const navigate = useNavigate();
  const { podium, title } = useGame()

  // Compute rank from sorted scores — resilient to missing backend rank field
  const topScore = podium[0]?.correct ?? -1;
  const getEffectiveRank = (player: LeaderboardEntry): number => {
    if (player.correct === topScore) return 1;
    const secondScore = podium.find(p => p && p.correct < topScore)?.correct;
    if (secondScore !== undefined && player.correct === secondScore) return 2;
    return 3;
  };

  const rankHeightPx: Record<number, string> = { 1: "176px", 2: "128px", 3: "96px" };
  const rankTrophyColor: Record<number, string> = {
    1: "hsl(42 90% 58%)",   // gold
    2: "hsl(220 15% 72%)",  // silver
    3: "hsl(25 70% 50%)",   // bronze
  };

  const winners = podium.filter(p => p && p.correct === topScore);
  const winnerText = winners.length <= 1
    ? `${podium[0]?.username ?? "No one"} wins!`
    : `${winners.map(w => w.username).join(" & ")} win!`;

  // Use classic 2nd–1st–3rd visual hierarchy only for exactly 3 players with distinct ranks.
  // For all other sizes, display in ranked order.
  const useClassicLayout = podium.length === 3
    && getEffectiveRank(podium[0]) !== getEffectiveRank(podium[1]);
  const podiumOrder = useClassicLayout
    ? [podium[1], podium[0], podium[2]]
    : podium;
  const podiumDelays = useClassicLayout
    ? ["0.3s", "0.1s", "0.5s"]
    : podiumOrder.map((_, i) => `${0.1 + i * 0.15}s`);

  return (
    <div className="relative h-screen flex flex-col overflow-hidden">
      <AnimatedBackground />

      {/* Header */}
      <header className="relative z-10 w-full max-w-3xl mx-auto px-6 pt-8 animate-fade-up">
        <div className="card-glass rounded-2xl px-6 py-4 text-center">
          <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
            Game Over
          </span>
          <h1 className="font-display text-2xl font-bold title-gradient leading-tight mt-1">
            {title}
          </h1>
        </div>
      </header>

      {/* Podium area */}
      <main className="relative z-10 flex-1 flex flex-col items-center justify-end pb-24 px-6">
        {/* Trophy + congrats */}
        <div
          className="flex flex-col items-center gap-2 mb-10 animate-fade-up"
          style={{ animationDelay: "0.2s", animationFillMode: "both" }}
        >
          <Trophy className="w-12 h-12" style={{ color: "hsl(42 90% 58%)" }} />
          <h2 className="font-display text-lg font-semibold text-foreground">
            {winnerText}
          </h2>
        </div>

        {/* Podium blocks — flex columns for ≤4 players, list for 5+ */}
        {podiumOrder.length <= 4 ? (
          <div className="flex items-end gap-4 w-full max-w-lg">
            {podiumOrder.map((player, i) => {
              if (!player) return <div key={i} className="flex-1" />;
              return (
                <div
                  key={player.username}
                  className="flex-1 flex flex-col items-center gap-2 animate-fade-up"
                  style={{ animationDelay: podiumDelays[i], animationFillMode: "both" }}
                >
                  {/* Name + score */}
                  <span className="text-sm font-semibold text-foreground truncate max-w-full">
                    {player.username}
                  </span>
                  <span
                    className="text-xs font-medium px-2 py-0.5 rounded-full"
                    style={{
                      background: `hsl(${player.color} / 0.2)`,
                      color: `hsl(${player.color})`,
                    }}
                  >
                    {player.correct} found
                  </span>

                  {/* Podium block */}
                  <div
                    className="w-full rounded-t-xl flex flex-col items-center justify-center gap-1 relative overflow-hidden"
                    style={{
                      height: rankHeightPx[getEffectiveRank(player)] ?? "96px",
                      background: `hsl(${player.color} / 0.18)`,
                      border: `1.5px solid hsl(${player.color} / 0.4)`,
                      boxShadow: `0 0 24px hsl(${player.color} / 0.2)`,
                    }}
                  >
                    <div
                      className="absolute inset-0 pointer-events-none"
                      style={{
                        background: `radial-gradient(ellipse at 50% 0%, hsl(${player.color} / 0.2) 0%, transparent 70%)`,
                      }}
                    />
                    <Trophy
                      className="relative w-6 h-6"
                      style={{ color: rankTrophyColor[getEffectiveRank(player)] }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="w-full max-w-md space-y-2">
            {podiumOrder.map((player, i) => (
              <div
                key={player.username}
                className="flex items-center gap-3 px-4 py-3 rounded-xl animate-fade-up"
                style={{
                  animationDelay: podiumDelays[i],
                  animationFillMode: "both",
                  background: `hsl(${player.color} / 0.12)`,
                  border: `1.5px solid hsl(${player.color} / 0.35)`,
                }}
              >
                <Trophy className="w-4 h-4 shrink-0" style={{ color: rankTrophyColor[getEffectiveRank(player)] ?? rankTrophyColor[3] }} />
                <span className="text-sm font-semibold text-foreground flex-1 truncate">
                  {player.username}
                </span>
                <span
                  className="text-xs font-medium px-2 py-0.5 rounded-full shrink-0"
                  style={{
                    background: `hsl(${player.color} / 0.2)`,
                    color: `hsl(${player.color})`,
                  }}
                >
                  {player.correct} found
                </span>
              </div>
            ))}
          </div>
        )}
      </main>

      {/* Play again */}
      <div className="relative z-10 w-full max-w-3xl mx-auto px-6 pb-8 animate-fade-up" style={{ animationDelay: "0.6s", animationFillMode: "both" }}>
        <button
          onClick={() => navigate("/")}
          className="btn-primary w-full text-center"
        >
          Play Again
        </button>
      </div>
    </div>
  );
};

export default Podium;
