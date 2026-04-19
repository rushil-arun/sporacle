/* eslint-disable react-refresh/only-export-components -- useGame is a required export */
import React, { createContext, useContext, useState } from 'react';
import type { LeaderboardEntry } from '@/types/types';

interface GameContextType {
  username: string;
  setUsername: (username: string) => void;
  code: string;
  setCode: (code: string) => void;
  wsUrl: string;
  setWsUrl: (url: string) => void;
  ws: WebSocket | null;
  setWs: (ws: WebSocket | null) => void;
  playerColor: string;
  setPlayerColor: (color: string) => void;
  podium: LeaderboardEntry[];
  setPodium: (pd : LeaderboardEntry[]) => void;
  title: string;
  setTitle: (t: string) => void;
  timeLeft: number;
  setTimeLeft: (t: any) => void;
  serverAddr: string;
  setServerAddr: (addr: string) => void;
  reset: () => void;
}

const GameContext = createContext<GameContextType | undefined>(undefined);

export const GameProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [username, setUsername] = useState('');
  const [code, setCode] = useState('');
  const [wsUrl, setWsUrl] = useState('');
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [playerColor, setPlayerColor] = useState('');
  const [podium, setPodium] = useState<LeaderboardEntry[]>([])
  const [title, setTitle] = useState("");
  const [timeLeft, setTimeLeft] = useState<number>(0);
  const [serverAddr, setServerAddr] = useState('');

  const reset = () => {
    if (ws) {
      ws.close();
      setWs(null);
    }
    setUsername('');
    setCode('');
    setWsUrl('');
    setPlayerColor('');
    setPodium([]);
    setServerAddr('');
  };

  return (
    <GameContext.Provider
      value={{
        username,
        setUsername,
        code,
        setCode,
        wsUrl,
        setWsUrl,
        ws,
        setWs,
        playerColor,
        setPlayerColor,
        podium,
        setPodium,
        title,
        setTitle,
        timeLeft,
        setTimeLeft,
        serverAddr,
        setServerAddr,
        reset,
      }}
    >
      {children}
    </GameContext.Provider>
  );
};

export const useGame = () => {
  const context = useContext(GameContext);
  if (!context) {
    throw new Error('useGame must be used within GameProvider');
  }
  return context;
};
