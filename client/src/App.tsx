import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Toaster } from '@/components/ui/toaster';
import { Toaster as Sonner } from '@/components/ui/sonner';
import { TooltipProvider } from '@/components/ui/tooltip';
import { GameProvider } from './context/GameContext';
import { Home } from './pages/Home';
import { Join } from './pages/Join';
import { Create } from './pages/Create';
import { CreateTemplate } from './pages/CreateTemplate';
import { Lobby } from './pages/Lobby';
import { Game } from './pages/Game'
import Podium from './pages/Podium';

function App() {
  return (
    <GameProvider>
      <TooltipProvider>
        <Toaster />
        <Sonner />
        <Router>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/join" element={<Join />} />
            <Route path="/create" element={<Create />} />
            <Route path="/create-template" element={<CreateTemplate />} />
            <Route path="/lobby" element={<Lobby />} />
            <Route path="/game" element={<Game />} />
            <Route path="/podium" element={<Podium />} />
          </Routes>
        </Router>
      </TooltipProvider>
    </GameProvider>
  );
}

export default App;
