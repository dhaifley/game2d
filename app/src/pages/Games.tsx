import React, { useState } from 'react';
import GamesTable from '../components/GamesTable';
import GameComponent from '../components/Game';
import { Game } from '../services/gameService';

const Games: React.FC = () => {
  const [selectedGame, setSelectedGame] = useState<Game | null>(null);

  const handleSelectGame = (game: Game) => {
    setSelectedGame(game);
  };

  const handleCloseGameView = () => {
    setSelectedGame(null);
  };

  return (
    <div className="games-container">
      <div className="games-header">
        <h1>Games</h1>
      </div>
      <div className="games-content">
      {selectedGame ? (
        <GameComponent game={selectedGame} onClose={handleCloseGameView} />
      ) : (
        <GamesTable onSelectGame={handleSelectGame} />
      )}
      </div>
    </div>
  );
};

export default Games;
