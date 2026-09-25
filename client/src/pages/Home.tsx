import { useNavigate } from 'react-router-dom';
import AnimatedBackground from '@/components/AnimatedBackground';

export const Home: React.FC = () => {
  const navigate = useNavigate();

  return (
    <div className="relative min-h-screen flex flex-col items-center justify-center overflow-hidden">
      <AnimatedBackground />

      <div className="relative z-10 flex flex-col items-center gap-8 px-4 animate-fade-up">
        <div className="text-center space-y-3">
          <h1
            className="title-gradient font-display text-7xl font-bold tracking-tight leading-none select-none"
            style={{ letterSpacing: '-0.03em' }}
          >
            Sporacle
          </h1>
          <p className="text-muted-foreground text-base font-body font-light tracking-wide max-w-xs mx-auto">
            Prove to your friends that you are an oracle.
          </p>
        </div>

        <div className="flex items-center gap-2 w-40">
          <div
            className="h-px flex-1"
            style={{ background: 'var(--gradient-brand)' }}
          />
          <div
            className="w-1.5 h-1.5 rounded-full"
            style={{ background: 'var(--gradient-brand)' }}
          />
          <div
            className="h-px flex-1"
            style={{ background: 'var(--gradient-brand)' }}
          />
        </div>

        <div className="flex flex-col gap-5 w-full max-w-[220px]">
          <div className="flex flex-col items-center gap-1.5">
            <button
              className="btn-primary text-sm w-full"
              onClick={() => navigate('/join')}
            >
              Join Lobby
            </button>
            <button
              className="text-muted-foreground hover:text-foreground transition-colors text-xs font-body underline underline-offset-4"
              onClick={() => navigate('/lobbies')}
            >
              Available Lobbies
            </button>
          </div>

          <div className="flex flex-col items-center gap-1.5">
            <button
              className="btn-secondary text-sm w-full"
              onClick={() => navigate('/create')}
            >
              Create Lobby
            </button>
            <button
              className="text-muted-foreground hover:text-foreground transition-colors text-xs font-body underline underline-offset-4"
              onClick={() => navigate('/create-template')}
            >
              Create Game Template
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
