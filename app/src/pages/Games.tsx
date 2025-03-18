import React, { useState, useRef } from 'react';
import GamesTable from '../components/GamesTable';
import GameComponent from '../components/Game';
import { Game } from '../services/gameService';

const Games: React.FC = () => {
  const [selectedGame, setSelectedGame] = useState<Game | null>(null);
  const gamesTableRef = useRef<{ loadGames: () => Promise<void> }>(null);

  const handleSelectGame = (game: Game) => {
    setSelectedGame(game);
  };

  const handleCloseGameView = () => {
    setSelectedGame(null);
  };

  const handleGameUpdated = async () => {
    // Refresh the games table data
    if (gamesTableRef.current) {
      await gamesTableRef.current.loadGames();
    }
  };

  return (
    <div className="games-container">
      <div className="games-header"></div>
      <div className="games-content">
      {selectedGame ? (
        <GameComponent 
          game={selectedGame} 
          onClose={handleCloseGameView} 
          onGameUpdated={handleGameUpdated}
        />
      ) : (
        <GamesTable 
          onSelectGame={handleSelectGame} 
          ref={gamesTableRef}
        />
      )}
      </div>
    </div>
  );
};

export default Games;
