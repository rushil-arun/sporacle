import { useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
import AnimatedBackground from '@/components/AnimatedBackground';
import { useGame } from '../context/GameContext';
import { useGetWsUrl } from '../hooks/useApi';

const CODE_LENGTH = 6;

export const Join: React.FC = () => {
  const navigate = useNavigate();
  const { username, setUsername, code, setCode, setWsUrl, setWs } = useGame();
  const { getWsUrl, loading, error } = useGetWsUrl();

  const [codeInputs, setCodeInputs] = useState<string[]>(
    code ? code.split('') : Array(CODE_LENGTH).fill('')
  );
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [connecting, setConnecting] = useState(false);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const handleCodeChange = (index: number, value: string) => {
    const char = value.replace(/[^a-zA-Z0-9]/g, '').slice(-1).toUpperCase();
    const next = [...codeInputs];
    next[index] = char;
    setCodeInputs(next);
    if (char && index < CODE_LENGTH - 1) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleCodeKeyDown = (
    index: number,
    e: React.KeyboardEvent<HTMLInputElement>
  ) => {
    if (e.key === 'Backspace' && !codeInputs[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault();
    const pasted = e.clipboardData
      .getData('text')
      .replace(/[^a-zA-Z0-9]/g, '')
      .toUpperCase()
      .slice(0, CODE_LENGTH);
    const next = Array(CODE_LENGTH).fill('');
    pasted.split('').forEach((char, i) => {
      next[i] = char;
    });
    setCodeInputs(next);
    const focusIndex = Math.min(pasted.length, CODE_LENGTH - 1);
    inputRefs.current[focusIndex]?.focus();
  };

  const setError = (msg: string) => {
    setSubmitError(msg);
    setConnecting(false);
  };

  const handleJoin = async () => {
    const fullCode = codeInputs.join('');

    if (!username.trim()) {
      setSubmitError('Please enter a username');
      return;
    }

    if (fullCode.length !== 6) {
      setSubmitError('Please enter all 6 characters of the game code');
      return;
    }

    setSubmitError(null);
    setConnecting(true);

    const baseWsUrl = await getWsUrl(username, fullCode);
    if (baseWsUrl) {
      setCode(fullCode);
      
      const fullWsUrl = `${(baseWsUrl)}`;
      setWsUrl(fullWsUrl);

      try {
        const ws = new WebSocket(fullWsUrl);

        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            console.log(data)
            if (data.type === 'error') {
              setError(data.message);
            } else if (data.type === 'success') {
              ws.onmessage = null
              setWs(ws);
              navigate('/lobby');
            }
          } catch {
            setError('Failed to connect.');
          }
        };

        ws.onerror = () => {
          setError('Failed to connect to game.');
        };

        ws.onclose = () => {
          if (!connecting) return;
        };
      } catch {
        setSubmitError('Failed to establish WebSocket connection');
        setConnecting(false);
      }
    } else {
      setSubmitError(error || 'Failed to join game');
      setConnecting(false);
    }
  };

  const isValid =
    username.trim().length > 0 && codeInputs.every((c) => c !== '');
  const isDisabled = connecting || !isValid || loading;

  return (
    <div className="relative min-h-screen flex flex-col items-center justify-center overflow-hidden">
      <AnimatedBackground />

      <div className="relative z-10 w-full max-w-sm px-4 animate-fade-up">
        <button
          onClick={() => navigate('/')}
          className="flex items-center gap-1.5 text-muted-foreground hover:text-foreground transition-colors text-sm mb-6 font-body"
        >
          <ArrowLeft className="w-4 h-4" />
          Back
        </button>

        <div className="card-glass rounded-2xl p-6 space-y-5">
          <div>
            <h2 className="font-display text-xl font-semibold text-foreground">
              Join a Game
            </h2>
            <p className="text-muted-foreground text-xs mt-0.5">
              Enter your name and the game code to join.
            </p>
          </div>

          <div>
            <label className="label-sporacle">Username</label>
            <input
              type="text"
              className="input-sporacle"
              placeholder="Your display name"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              maxLength={20}
              autoFocus
              disabled={connecting}
            />
          </div>

          <div>
            <label className="label-sporacle">Game Code</label>
            <div className="flex gap-2 justify-between">
              {codeInputs.map((char, i) => (
                <input
                  key={i}
                  ref={(el) => {
                    inputRefs.current[i] = el;
                  }}
                  type="text"
                  inputMode="text"
                  maxLength={1}
                  value={char}
                  className={`code-box flex-1 ${char ? 'filled' : ''}`}
                  onChange={(e) => handleCodeChange(i, e.target.value)}
                  onKeyDown={(e) => handleCodeKeyDown(i, e)}
                  onPaste={i === 0 ? handlePaste : undefined}
                  disabled={connecting}
                />
              ))}
            </div>
          </div>

          {(submitError || error) && (
            <div className="p-3 rounded-lg bg-destructive/20 border border-destructive/30 text-destructive text-sm">
              {submitError || error}
            </div>
          )}

          <button
            className="btn-primary w-full text-sm"
            disabled={isDisabled}
            style={
              isDisabled
                ? {
                    opacity: 0.45,
                    cursor: 'not-allowed',
                    transform: 'none',
                    boxShadow: 'none',
                  }
                : {}
            }
            onClick={handleJoin}
          >
            {connecting ? 'Connecting...' : loading ? 'Joining...' : 'Join Game'}
          </button>
        </div>
      </div>
    </div>
  );
};
