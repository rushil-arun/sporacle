import { useNavigate } from "react-router-dom";
import AnimatedBackground from "@/components/AnimatedBackground";
import { Trophy } from "lucide-react";
import { useGame } from '../context/GameContext';



const Podium = () => {
  const navigate = useNavigate();
  const { podium, title } = useGame()

  // Podium order: 2nd, 1st, 3rd for visual layout
  const podiumOrder = [podium[1], podium[0], podium[2]]
  const podiumHeights = ["h-32", "h-44", "h-24"];
  const podiumDelays = ["0.3s", "0.1s", "0.5s"];
  const placeLabels = ["2nd", "1st", "3rd"];
  const trophyColors = [
    "hsl(220 15% 72%)", // silver
    "hsl(42 90% 58%)",  // gold
    "hsl(25 70% 50%)",  // bronze
  ];

  console.log(podiumOrder)

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
            {podium[0]?.username ?? "No one"} wins!
          </h2>
        </div>

        {/* Podium blocks */}
        <div className="flex items-end gap-4 w-full max-w-md">
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
                  className={`w-full ${podiumHeights[i]} rounded-t-xl flex flex-col items-center justify-center gap-1 relative overflow-hidden`}
                  style={{
                    background: `hsl(${player.color} / 0.18)`,
                    border: `1.5px solid hsl(${player.color} / 0.4)`,
                    boxShadow: `0 0 24px hsl(${player.color} / 0.2)`,
                  }}
                >
                  {/* Radial glow */}
                  <div
                    className="absolute inset-0 pointer-events-none"
                    style={{
                      background: `radial-gradient(ellipse at 50% 0%, hsl(${player.color} / 0.2) 0%, transparent 70%)`,
                    }}
                  />
                  <Trophy
                    className="relative w-6 h-6"
                    style={{ color: trophyColors[i] }}
                  />
                  <span
                    className="relative font-display text-xl font-bold"
                    style={{ color: `hsl(${player.color})` }}
                  >
                    {placeLabels[i]}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
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
