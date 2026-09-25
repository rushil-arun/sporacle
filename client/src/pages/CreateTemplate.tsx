import { useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, X } from 'lucide-react';
import AnimatedBackground from '@/components/AnimatedBackground';
import { AutocompleteInput } from '@/components/AutocompleteInput';
import {
  useCreateCategory,
  useSearchBroadCategories,
  useSearchNarrowCategories,
} from '../hooks/useApi';

export const CreateTemplate: React.FC = () => {
  const navigate = useNavigate();
  const { searchBroadCategories } = useSearchBroadCategories();
  const { searchNarrowCategories } = useSearchNarrowCategories();
  const { createCategory, loading: creating } = useCreateCategory();

  const [broadCategory, setBroadCategory] = useState('');
  const [narrowCategory, setNarrowCategory] = useState('');
  const [itemInput, setItemInput] = useState('');
  const [items, setItems] = useState<string[]>([]);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const itemInputRef = useRef<HTMLInputElement>(null);

  const commitItem = () => {
    const trimmed = itemInput.trim();
    if (!trimmed) return;
    if (items.some((item) => item.toLowerCase() === trimmed.toLowerCase())) {
      return;
    }
    setItems((prev) => [...prev, trimmed]);
    setItemInput('');
  };

  const handleItemKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      commitItem();
    }
  };

  const editItem = (item: string) => {
    setItems((prev) => prev.filter((i) => i !== item));
    setItemInput(item);
    itemInputRef.current?.focus();
  };

  const removeItem = (e: React.MouseEvent, item: string) => {
    e.stopPropagation();
    setItems((prev) => prev.filter((i) => i !== item));
  };

  const handleConfirm = async () => {
    setSubmitError(null);
    setSuccess(false);

    // Auto-commit any text still sitting in the item box.
    const pending = itemInput.trim();
    const finalItems = [...items];
    if (pending && !finalItems.some((i) => i.toLowerCase() === pending.toLowerCase())) {
      finalItems.push(pending);
    }

    if (!broadCategory.trim() || !narrowCategory.trim()) {
      setSubmitError('Please fill in both category fields');
      return;
    }
    if (finalItems.length === 0) {
      setSubmitError('Please add at least one item');
      return;
    }

    const { data, error } = await createCategory(broadCategory.trim(), narrowCategory.trim(), finalItems);
    if (data) {
      setBroadCategory('');
      setNarrowCategory('');
      setItems([]);
      setItemInput('');
      setSuccess(true);
    } else {
      setSubmitError(error || 'Failed to create category');
    }
  };

  const pendingTrimmed = itemInput.trim();
  const totalItemCount = items.length + (pendingTrimmed ? 1 : 0);
  const isValid = broadCategory.trim() !== '' && narrowCategory.trim() !== '' && totalItemCount > 0;

  return (
    <div className="relative min-h-screen flex flex-col items-center justify-center overflow-hidden">
      <AnimatedBackground />

      <div className="relative z-10 w-full max-w-sm px-4 animate-fade-up">
        <button
          onClick={() => navigate('/create')}
          className="flex items-center gap-1.5 text-muted-foreground hover:text-foreground transition-colors text-sm mb-6 font-body"
        >
          <ArrowLeft className="w-4 h-4" />
          Back
        </button>

        <div className="card-glass rounded-2xl p-6 space-y-5">
          <div>
            <h2 className="font-display text-xl font-semibold text-foreground">
              Create a Game Template
            </h2>
            <p className="text-muted-foreground text-xs mt-0.5">
              Add a category and its items so others can play it.
            </p>
          </div>

          <AutocompleteInput
            label="Broad Category"
            value={broadCategory}
            onChange={(v) => {
              setBroadCategory(v);
              setSuccess(false);
            }}
            fetchSuggestions={searchBroadCategories}
            placeholder="e.g. Sports"
            disabled={creating}
          />

          <AutocompleteInput
            label="Narrow Category"
            value={narrowCategory}
            onChange={(v) => {
              setNarrowCategory(v);
              setSuccess(false);
            }}
            fetchSuggestions={searchNarrowCategories}
            placeholder="e.g. NHL Teams"
            disabled={creating}
          />

          <div>
            <label className="label-sporacle">Items</label>
            <input
              ref={itemInputRef}
              className="input-sporacle"
              type="text"
              value={itemInput}
              onChange={(e) => setItemInput(e.target.value)}
              onKeyDown={handleItemKeyDown}
              placeholder="Type an item and press Enter…"
              disabled={creating}
            />
            {items.length > 0 && (
              <div className="flex flex-wrap gap-2 mt-3">
                {items.map((item) => (
                  <button
                    key={item}
                    type="button"
                    onClick={() => editItem(item)}
                    className="flex items-center gap-1.5 bg-muted hover:bg-muted/70 text-foreground text-xs rounded-full pl-3 pr-2 py-1.5 transition-colors"
                  >
                    <span>{item}</span>
                    <X
                      className="w-3 h-3 text-muted-foreground hover:text-foreground"
                      onClick={(e) => removeItem(e, item)}
                    />
                  </button>
                ))}
              </div>
            )}
          </div>

          {submitError && (
            <div className="p-3 rounded-lg bg-destructive/20 border border-destructive/30 text-destructive text-sm">
              {submitError}
            </div>
          )}

          {success && (
            <div className="p-3 rounded-lg bg-accent/20 border border-accent/30 text-accent text-sm">
              Category created! Add another, or head back.
            </div>
          )}

          <button
            className="btn-secondary w-full text-sm"
            disabled={creating || !isValid}
            style={
              creating || !isValid
                ? {
                    opacity: 0.45,
                    cursor: 'not-allowed',
                    transform: 'none',
                    boxShadow: 'none',
                  }
                : {}
            }
            onClick={handleConfirm}
          >
            {creating ? 'Creating...' : 'Confirm'}
          </button>
        </div>
      </div>
    </div>
  );
};
