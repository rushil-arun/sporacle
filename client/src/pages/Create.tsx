import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
import AnimatedBackground from '@/components/AnimatedBackground';
import { useGame } from '../context/GameContext';
import { useCreateGame, useTriviaFiles, useTriviaKeys } from '../hooks/useApi';

const capitalizeWords = (str: string): string => {
  return str
    .replace(/\.json$/, '')
    .split(/[\s_-]+/)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
};

export const Create: React.FC = () => {
  const navigate = useNavigate();
  const { setCode, setTimeLeft, setServerAddr } = useGame();
  const {
    createGame,
    loading: creatingGame,
    error: createError,
  } = useCreateGame();
  const { fetchFiles, loading: loadingFiles, error: filesError } =
    useTriviaFiles();
  const { fetchKeys, loading: loadingKeys, error: keysError } =
    useTriviaKeys();

  const [categories, setCategories] = useState<string[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [difficulties, setDifficulties] = useState<string[]>([]);
  const [selectedDifficulty, setSelectedDifficulty] = useState<string>('');
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [lobbyTime, setLobbyTime] = useState("");
  const [gameTime, setGameTime] = useState("");

  useEffect(() => {
    const loadCategories = async () => {
      const files = await fetchFiles();
      if (files) {
        setCategories(files);
        setSelectedCategory(files[0] || '');
      }
    };
    loadCategories();
    // eslint-disable-next-line react-hooks/exhaustive-deps -- run only on mount
  }, []);

  useEffect(() => {
    const loadDifficulties = async () => {
      if (selectedCategory) {
        const keys = await fetchKeys(selectedCategory);
        if (keys) {
          setDifficulties(keys);
          setSelectedDifficulty(keys[0] || '');
        }
      } else {
        setDifficulties([]);
        setSelectedDifficulty('');
      }
    };
    loadDifficulties();
    // eslint-disable-next-line react-hooks/exhaustive-deps -- fetchKeys is stable
  }, [selectedCategory]);

  const handleCategoryChange = (value: string) => {
    setSelectedCategory(value);
    setSelectedDifficulty('');
  };

  const handleCreate = async () => {
    if (!selectedDifficulty) {
      setSubmitError('Please select a category and difficulty');
      return;
    }
    const result = await createGame(selectedDifficulty, Number(lobbyTime), Number(gameTime));
    if (result) {
      setCode(result.code);
      setServerAddr(result.serverAddr);
      setTimeLeft(0);
      navigate('/join');
    } else {
      setSubmitError(createError || 'Failed to create game');
    }
  };

  const timesValid =
    lobbyTime !== "" && Number(lobbyTime) >= 10 &&
    gameTime !== "" && Number(gameTime) >= 10;
  const isValid = selectedCategory !== "" && selectedDifficulty !== "" && timesValid;

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
              Create a Game
            </h2>
            <p className="text-muted-foreground text-xs mt-0.5">
              Choose a category and difficulty to generate your game.
            </p>
          </div>

          <div>
            <label className="label-sporacle">Category</label>
            <select
              className="select-sporacle"
              value={selectedCategory}
              onChange={(e) => handleCategoryChange(e.target.value)}
              disabled={loadingFiles}
            >
              <option value="" disabled>
                Select a category…
              </option>
              {loadingFiles ? (
                <option value="">Loading...</option>
              ) : filesError ? (
                <option value="">Error loading categories</option>
              ) : (
                categories.map((cat) => (
                  <option key={cat} value={cat}>
                    {capitalizeWords(cat)}
                  </option>
                ))
              )}
            </select>
          </div>

          {selectedCategory && (
            <div
              className="animate-fade-up"
              style={{ animationDuration: '0.25s' }}
            >
              <label className="label-sporacle">Difficulty</label>
              <select
                className="select-sporacle"
                value={selectedDifficulty}
                onChange={(e) => setSelectedDifficulty(e.target.value)}
                disabled={loadingKeys}
              >
                <option value="" disabled>
                  Select a difficulty…
                </option>
                {loadingKeys ? (
                  <option value="">Loading...</option>
                ) : keysError ? (
                  <option value="">Error loading difficulties</option>
                ) : (
                  difficulties.map((d) => (
                    <option key={d} value={d}>
                      {capitalizeWords(d)}
                    </option>
                  ))
                )}
              </select>
            </div>
          )}

          {(submitError || createError) && (
            <div className="p-3 rounded-lg bg-destructive/20 border border-destructive/30 text-destructive text-sm">
              {submitError || createError}
            </div>
          )}

          {/* Time settings — only shown once difficulty selected */}
          {selectedDifficulty && (
            <div className="animate-fade-up grid grid-cols-2 gap-3" style={{ animationDuration: "0.25s" }}>
              <div>
                <label className="label-sporacle">Lobby Time (sec)</label>
                <input
                  type="number"
                  min={10}
                  className="input-sporacle"
                  placeholder="e.g. 30"
                  value={lobbyTime}
                  onChange={(e) => setLobbyTime(e.target.value)}
                />
                {lobbyTime !== "" && Number(lobbyTime) < 10 && (
                  <p className="text-destructive text-xs mt-1">Must be ≥ 10</p>
                )}
              </div>
              <div>
                <label className="label-sporacle">Game Time (sec)</label>
                <input
                  type="number"
                  min={10}
                  className="input-sporacle"
                  placeholder="e.g. 120"
                  value={gameTime}
                  onChange={(e) => setGameTime(e.target.value)}
                />
                {gameTime !== "" && Number(gameTime) < 10 && (
                  <p className="text-destructive text-xs mt-1">Must be ≥ 10</p>
                )}
              </div>
            </div>
          )}

          
          <button
            className="btn-secondary w-full text-sm"
            disabled={creatingGame || !isValid}
            style={
              creatingGame || !isValid
                ? {
                    opacity: 0.45,
                    cursor: 'not-allowed',
                    transform: 'none',
                    boxShadow: 'none',
                  }
                : {}
            }
            onClick={handleCreate}
          >
            {creatingGame ? 'Creating...' : 'Create Game'}
          </button>
        </div>
      </div>
    </div>
  );
};
