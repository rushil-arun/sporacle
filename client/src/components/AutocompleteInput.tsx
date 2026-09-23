import { useEffect, useRef, useState } from 'react';

interface AutocompleteInputProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  fetchSuggestions: (query: string) => Promise<string[] | null>;
  placeholder?: string;
  disabled?: boolean;
}

// AutocompleteInput is a free-solo text input: the user can pick a suggestion or
// keep typing something new. Suggestions are debounced against fetchSuggestions.
export const AutocompleteInput: React.FC<AutocompleteInputProps> = ({
  label,
  value,
  onChange,
  fetchSuggestions,
  placeholder,
  disabled,
}) => {
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [open, setOpen] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const blurTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  useEffect(() => {
    clearTimeout(debounceRef.current);
    if (value.trim() === '') {
      setSuggestions([]);
      return;
    }
    debounceRef.current = setTimeout(async () => {
      const results = await fetchSuggestions(value.trim());
      setSuggestions(results ?? []);
    }, 250);
    return () => clearTimeout(debounceRef.current);
    // eslint-disable-next-line react-hooks/exhaustive-deps -- fetchSuggestions is stable per page
  }, [value]);

  useEffect(() => () => clearTimeout(blurTimeoutRef.current), []);

  const handleSelect = (suggestion: string) => {
    onChange(suggestion);
    setOpen(false);
  };

  return (
    <div className="relative">
      <label className="label-sporacle">{label}</label>
      <input
        className="input-sporacle"
        type="text"
        value={value}
        placeholder={placeholder}
        disabled={disabled}
        onChange={(e) => {
          onChange(e.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
        onBlur={() => {
          // Delay so a click on a suggestion registers before the list unmounts.
          blurTimeoutRef.current = setTimeout(() => setOpen(false), 150);
        }}
      />
      {open && suggestions.length > 0 && (
        <ul className="absolute z-10 mt-1 w-full max-h-48 overflow-y-auto rounded-xl border border-border bg-popover shadow-lg py-1">
          {suggestions.map((s) => (
            <li key={s}>
              <button
                type="button"
                className="w-full text-left px-4 py-2 text-sm text-foreground hover:bg-muted transition-colors"
                onClick={() => handleSelect(s)}
              >
                {s}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};
