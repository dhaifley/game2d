import React from 'react';
import { Game as GameType } from '../services/gameService';

interface GameProps {
  game: GameType;
  onClose: () => void;
}

const Game: React.FC<GameProps> = ({ game, onClose }) => {
  return (
    <div className="game-details-container">
      <div className="game-details-header">
        <h2>{game.name}</h2>
        <button className="close-button" onClick={onClose}>Close</button>
      </div>
      <div className="game-details-content">
        <p>ID: {game.id}</p>
        <p>Name: {game.name}</p>
        <p>Version: {game.version || 'N/A'}</p>
        <p>Status: {game.status || 'N/A'}</p>
        <p>Updated At: {game.updated_at ? new Date(game.updated_at * 1000).toLocaleString() : 'N/A'}</p>
        <p>Updated By: {game.updated_by || 'N/A'}</p>
      </div>
    </div>
  );
};

export default Game;
