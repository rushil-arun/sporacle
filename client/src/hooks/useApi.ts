import axios, { AxiosError } from 'axios';
import { useState } from 'react';
import { pickRandomServer } from '@/lib/serverPool';

interface CreateGameResponse {
  code: string;
  serverAddr: string;
}

interface CreateGameError {
  error?: string;
}

export const useCreateGame = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const createGame = async (
    title: string,
    gameTime: number
  ): Promise<{ code: string; serverAddr: string } | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.post<CreateGameResponse>(
        `${pickRandomServer()}/create-game`,
        { title, gameTime },
        { headers: { 'Content-Type': 'application/json' } }
      );
      return { code: response.data.code, serverAddr: response.data.serverAddr };
    } catch (err) {
      const axiosError = err as AxiosError<CreateGameError>;
      const errorMessage =
        axiosError.response?.data?.error ||
        'Failed to create lobby';
      setError(errorMessage);
      return null;
    } finally {
      setLoading(false);
    }
  };

  return { createGame, loading, error };
};

interface GetWsUrlResponse {
  url: string;
}

interface GetWsUrlError {
  error?: string;
}

export const useGetWsUrl = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const getWsUrl = async (
    username: string,
    code: string
  ): Promise<string | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.get<GetWsUrlResponse>(
        `${pickRandomServer()}/get-ws-url`,
        {
          params: { "username": username, "code": code },
          headers: { 'Content-Type': 'application/json' },
        }
      );
      return response.data.url;
    } catch (err) {
      const axiosError = err as AxiosError<GetWsUrlError>;
      const errorMessage =
        axiosError.response?.data?.error ||
        axiosError.message ||
        'Failed to get WebSocket URL';
      setError(errorMessage);
      return null;
    } finally {
      setLoading(false);
    }
  };

  return { getWsUrl, loading, error };
};

export interface Lobby {
  code: string;
  title: string;
  creator: string;
}

interface LobbiesResponse {
  lobbies: Lobby[];
}

export const useLobbies = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchLobbies = async (): Promise<Lobby[] | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.get<LobbiesResponse>(
        `${pickRandomServer()}/lobbies`,
        { headers: { 'Content-Type': 'application/json' } }
      );
      return response.data.lobbies;
    } catch (err) {
      const axiosError = err as AxiosError;
      const errorMessage = axiosError.message || 'Failed to fetch lobbies';
      setError(errorMessage);
      return null;
    } finally {
      setLoading(false);
    }
  };

  return { fetchLobbies, loading, error };
};

export const useTriviaFiles = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchFiles = async (): Promise<string[] | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.get<string[]>(
        `${pickRandomServer()}/trivia/files`,
        { headers: { 'Content-Type': 'application/json' } }
      );
      return response.data;
    } catch (err) {
      const axiosError = err as AxiosError;
      const errorMessage = axiosError.message || 'Failed to fetch trivia files';
      setError(errorMessage);
      return null;
    } finally {
      setLoading(false);
    }
  };

  return { fetchFiles, loading, error };
};

export const useTriviaKeys = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchKeys = async (file: string): Promise<string[] | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.get<string[]>(
        `${pickRandomServer()}/trivia/keys`,
        {
          params: { file },
          headers: { 'Content-Type': 'application/json' },
        }
      );
      return response.data;
    } catch (err) {
      const axiosError = err as AxiosError;
      const errorMessage = axiosError.message || 'Failed to fetch trivia keys';
      setError(errorMessage);
      return null;
    } finally {
      setLoading(false);
    }
  };

  return { fetchKeys, loading, error };
};

export const useSearchBroadCategories = () => {
  const searchBroadCategories = async (query: string): Promise<string[] | null> => {
    try {
      const response = await axios.get<string[]>(
        `${pickRandomServer()}/trivia/broad-categories`,
        { params: { q: query } }
      );
      return response.data;
    } catch {
      return null;
    }
  };

  return { searchBroadCategories };
};

export const useSearchNarrowCategories = () => {
  const searchNarrowCategories = async (query: string): Promise<string[] | null> => {
    try {
      const response = await axios.get<string[]>(
        `${pickRandomServer()}/trivia/narrow-categories`,
        { params: { q: query } }
      );
      return response.data;
    } catch {
      return null;
    }
  };

  return { searchNarrowCategories };
};

interface CreateCategoryResponse {
  broadCategory: string;
  narrowCategory: string;
  items: string[];
}

interface CreateCategoryError {
  error?: string;
}

export const useCreateCategory = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Returns the error message directly (not just via the `error` state) since a
  // caller awaiting this promise runs before the state update above is applied,
  // so reading the hook's `error` state right after awaiting would be stale.
  const createCategory = async (
    broadCategory: string,
    narrowCategory: string,
    items: string[]
  ): Promise<{ data: CreateCategoryResponse | null; error: string | null }> => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.post<CreateCategoryResponse>(
        `${pickRandomServer()}/trivia/categories`,
        { broadCategory, narrowCategory, items },
        { headers: { 'Content-Type': 'application/json' } }
      );
      return { data: response.data, error: null };
    } catch (err) {
      const axiosError = err as AxiosError<CreateCategoryError>;
      const errorMessage =
        axiosError.response?.data?.error || 'Failed to create category';
      setError(errorMessage);
      return { data: null, error: errorMessage };
    } finally {
      setLoading(false);
    }
  };

  return { createCategory, loading, error };
};
