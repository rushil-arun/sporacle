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
    lobbyTime: number,
    gameTime: number
  ): Promise<{ code: string; serverAddr: string } | null> => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.post<CreateGameResponse>(
        `${pickRandomServer()}/create-game`,
        { title, lobbyTime, gameTime },
        { headers: { 'Content-Type': 'application/json' } }
      );
      return { code: response.data.code, serverAddr: response.data.serverAddr };
    } catch (err) {
      const axiosError = err as AxiosError<CreateGameError>;
      const errorMessage =
        axiosError.response?.data?.error ||
        'Failed to create game';
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
