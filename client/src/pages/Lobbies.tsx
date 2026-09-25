import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, RefreshCw, Users } from 'lucide-react';
import AnimatedBackground from '@/components/AnimatedBackground';
import { useGame } from '../context/GameContext';
import { useLobbies, type Lobby } from '../hooks/useApi';

const POLL_INTERVAL_MS = 5000;

const formatTimeLeft = (seconds: number): string => {
  if (seconds <= 0) return '0:00';
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${s.toString().padStart(2, '0')}`;
};

export const Lobbies: React.FC = () => {
  const navigate = useNavigate();
  const { setCode } = useGame();
  const { fetchLobbies, loading, error } = useLobbies();

  const [lobbies, setLobbies] = useState<Lobby[]>([]);
  const [hasLoaded, setHasLoaded] = useState(false);

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      const result = await fetchLobbies();
      if (!cancelled && result) {
        setLobbies(result);
        setHasLoaded(true);
      }
    };

    load();
    const interval = setInterval(load, POLL_INTERVAL_MS);
    return () => {
      cancelled = true;
      clearInterval(interval);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- fetchLobbies is stable
  }, []);

  // Tick the displayed countdowns down locally every second between polls,
  // so they don't sit frozen for up to POLL_INTERVAL_MS at a time.
  useEffect(() => {
    const tick = setInterval(() => {
      setLobbies((prev) =>
        prev.map((lobby) => ({
          ...lobby,
          timeLeft: Math.max(0, lobby.timeLeft - 1),
        }))
      );
    }, 1000);
    return () => clearInterval(tick);
  }, []);

  const handleJoin = (code: string) => {
    setCode(code);
    navigate('/join', { state: { from: 'lobbies' } });
  };

  return (
    <div className="relative min-h-screen flex flex-col items-center justify-center overflow-hidden">
      <AnimatedBackground />

      <div className="relative z-10 w-full max-w-sm px-4 animate-fade-up">
        <button
          onClick={() => navigate('/join')}
          className="flex items-center gap-1.5 text-muted-foreground hover:text-foreground transition-colors text-sm mb-6 font-body"
        >
          <ArrowLeft className="w-4 h-4" />
          Back
        </button>

        <div className="card-glass rounded-2xl p-6 space-y-4">
          <div className="flex items-start justify-between gap-2">
            <div>
              <h2 className="font-display text-xl font-semibold text-foreground">
                Available Lobbies
              </h2>
              <p className="text-muted-foreground text-xs mt-0.5">
                Games that haven't started yet. Jump into one below.
              </p>
            </div>
            <button
              onClick={() => fetchLobbies().then((r) => r && setLobbies(r))}
              className="text-muted-foreground hover:text-foreground transition-colors mt-1"
              aria-label="Refresh"
              disabled={loading}
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>

          {error && (
            <div className="p-3 rounded-lg bg-destructive/20 border border-destructive/30 text-destructive text-sm">
              {error}
            </div>
          )}

          {!error && hasLoaded && lobbies.length === 0 && (
            <div className="text-center py-8 text-muted-foreground text-sm">
              No open lobbies right now.
              <br />
              Why not create one?
            </div>
          )}

          {!error && !hasLoaded && (
            <div className="text-center py-8 text-muted-foreground text-sm">
              Loading lobbies...
            </div>
          )}

          {lobbies.length > 0 && (
            <div className="space-y-2 max-h-80 overflow-y-auto pr-1">
              {lobbies.map((lobby) => (
                <div
                  key={lobby.code}
                  className="flex items-center justify-between gap-3 p-3 rounded-lg bg-white/5 border border-white/10"
                >
                  <div className="min-w-0">
                    <p className="font-body text-sm font-medium text-foreground truncate">
                      {lobby.title}
                    </p>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                      <Users className="w-3 h-3" />
                      <span className="truncate">
                        {lobby.creator ? lobby.creator : 'Waiting for host'}
                      </span>
                      <span>·</span>
                      <span>{formatTimeLeft(lobby.timeLeft)} left</span>
                    </div>
                  </div>
                  <button
                    className="btn-secondary text-xs px-3 py-1.5 shrink-0"
                    onClick={() => handleJoin(lobby.code)}
                  >
                    Join
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
