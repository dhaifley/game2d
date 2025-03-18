import React, { useEffect, useState } from 'react';
import { Game as GameType } from '../services/gameService';
import avatarImage from '../assets/avatar.png';

interface GameProps {
  game: GameType;
  onClose: () => void;
}

const Game: React.FC<GameProps> = ({ game, onClose }) => {
  // Determine icon source: use base64 SVG from game.icon or fallback to avatar.png
  const iconSrc = game.icon ? `data:image/svg+xml;base64,${game.icon}` : avatarImage;
  return (
    <div className="game-details-container">
      <div className="game-details-header">
        <h2>{game.name}</h2>
        <div className="game-buttons">
          <button className="delete-button" onClick={onClose}>Delete</button>
          <button className="copy-button" onClick={onClose}>Copy</button>
          <button className="edit-button" onClick={onClose}>Edit</button>
          <button className="close-button" onClick={onClose}>Close</button>
        </div>
      </div>

      <div className="game-details-content">
        <div className="game-info-layout">
          <div className="game-icon-container">
            <img 
              src={iconSrc} 
              alt={`${game.name} icon`} 
              className="game-icon" 
              width="128" 
              height="128"
            />
          </div>
          <div className="game-info-fields">
            <div className="game-field-container">
              <label htmlFor='id'>ID: </label>
              <input
                type="text"
                id="id"
                value={game.id}
                className="readonly-input"
                readOnly
              />
            </div>
            <div className="game-field-container">
              <label htmlFor='name'>Name: </label>
              <input
                type="text"
                id="name"
                value={game.name}
                placeholder="Enter Name"
                className="name-input"
                readOnly
              />
            </div>
            <div className="game-field-container">
              <label htmlFor='status'>Status: </label>
              <input
                type="text"
                id="status"
                value={game.status || ''}
                className="readonly-input"
                readOnly
              />
            </div>
          </div>
        </div>

        {game.status_data && (
          <div className="status-data-container">
            <h3>Status Data:</h3>
            <pre className="status-data-json">
              {JSON.stringify(game.status_data, null, 2)}
            </pre>
          </div>
        )}

        <div className="client-container">
          {game.id && <GameIframe gameId={game.id} />}
        </div>

        <div className="ai-field-container">
          <div className="response-container">
            <textarea 
              id="response"
              className="response-textarea"
              readOnly
              value={game.response || ''}
            />
          </div>
          <div className="prompt-container">
            <textarea
              id="prompt"
              value={game.ai_data?.prompt || ''}
              placeholder="Enter prompt"
              className="prompt-textarea"
            />
            <button className="prompt-button" onClick={onClose}>Prompt</button>
            <button className={game.previous_id ? "undo-button" : "undo-button-disabled"} onClick={onClose}>Undo</button>
          </div>
        </div>

        <div className="game-field-container">
          <label htmlFor='description'>Description: </label>
          <textarea 
            id="description"
            className="description-textarea"
            placeholder="Enter Description"
            readOnly
            value={game.description || ''}
          />
        </div>

        <div className="game-field-container">
          <label htmlFor='tags'>Tags: </label>
          <input
            type="text"
            id="tags"
            value={game.tags?.join(', ') || ''}
            placeholder="Enter Tags"
            className="tags-input"
            readOnly
          />
        </div>

        <div className="game-additional-fields">
          <div className="game-field-container">
            <label htmlFor='source'>Source: </label>
            <input
              type="text"
              id="source"
              value={game.source || ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="game-field-container">
            <label htmlFor='created_at'>Created At: </label>
            <input
              type="text"
              id="created_at"
              value={game.created_at ? new Date(game.created_at * 1000).toLocaleString() : ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="game-field-container">
            <label htmlFor='created_by'>Created By: </label>
            <input
              type="text"
              id="created_by"
              value={game.created_by || ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="game-field-container">
            <label htmlFor='updated_at'>Updated At: </label>
            <input
              type="text"
              id="updated_at"
              value={game.updated_at ? new Date(game.updated_at * 1000).toLocaleString() : ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="game-field-container">
            <label htmlFor='updated_by'>Updated By: </label>
            <input
              type="text"
              id="updated_by"
              value={game.updated_by || ''}
              className="readonly-input"
              readOnly
            />
          </div>
        </div>
      </div>
    </div>
  );
};

// Component for the game iframe that handles token retrieval and refresh logic
const GameIframe: React.FC<{ gameId: string }> = ({ gameId }) => {
  const [apiToken, setApiToken] = useState<string>('');
  
  // Get API token from localStorage on component mount and when gameId changes
  useEffect(() => {
    const token = localStorage.getItem('token') || '';
    setApiToken(token);
  }, [gameId]);

  // Build the iframe src URL with query parameters
  const iframeSrc = apiToken ? 
    `/client?game_id=${gameId}&api_url=${window.location.origin}/api/v1&api_token=${apiToken}` : '';

  // Only render the iframe when we have an API token
  return apiToken ? (
    <iframe 
      src={iframeSrc}
      title="Game Client"
      width="640px"
      height="480px"
      frameBorder="0"
      allowFullScreen
    />
  ) : null;
};

export default Game;
