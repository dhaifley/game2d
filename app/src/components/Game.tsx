import React, { useEffect, useState } from 'react';
import { Game as GameType, copyGame, deleteGame, updateGame } from '../services/gameService';
import avatarImage from '../assets/avatar.png';
import Modal from './Modal';
import axios from 'axios';

interface GameProps {
  game: GameType;
  onClose: () => void;
  onGameUpdated: () => Promise<void>; // Function to refresh games list
}

const Game: React.FC<GameProps> = ({ game, onClose, onGameUpdated }) => {
  // Determine icon source: use base64 SVG from game.icon or fallback to avatar.png
  const iconSrc = game.icon ? `data:image/svg+xml;base64,${game.icon}` : avatarImage;
  
  // Edit mode state
  const [isEditMode, setIsEditMode] = useState(false);
  const [editedName, setEditedName] = useState(game.name);
  const [editedDescription, setEditedDescription] = useState(game.description || '');
  const [editedTags, setEditedTags] = useState(game.tags?.join(', ') || '');
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  
  // Copy modal state
  const [isCopyModalOpen, setIsCopyModalOpen] = useState(false);
  const [copyName, setCopyName] = useState('');
  const [copyError, setCopyError] = useState<string | null>(null);
  const [isCopying, setIsCopying] = useState(false);
  
  // Delete modal state
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  
  // State to keep track of the current game data
  const [currentGame, setCurrentGame] = useState<GameType>(game);
  
  // Update form fields when game prop changes
  useEffect(() => {
    setCurrentGame(game);
    setEditedName(game.name);
    setEditedDescription(game.description || '');
    setEditedTags(game.tags?.join(', ') || '');
  }, [game]);
  
  // Handle entering edit mode
  const handleEditClick = () => {
    setIsEditMode(true);
    setSaveError(null);
  };
  
  // Handle canceling edit mode (restore original values)
  const handleCancelEdit = () => {
    setIsEditMode(false);
    setEditedName(currentGame.name);
    setEditedDescription(currentGame.description || '');
    setEditedTags(currentGame.tags?.join(', ') || '');
    setSaveError(null);
  };
  
  // Handle saving edited game
  const handleSaveEdit = async () => {
    if (!editedName.trim()) {
      setSaveError('Game name is required');
      return;
    }
    
    try {
      setIsSaving(true);
      setSaveError(null);
      
      // Prepare the update payload
      const updates: Partial<GameType> = {
        name: editedName.trim(),
        description: editedDescription.trim(),
        tags: editedTags.trim() ? editedTags.split(',').map(tag => tag.trim()) : []
      };
      
      // Call the API to update the game
      const updatedGame = await updateGame(game.id, updates);
      
      // Exit edit mode
      setIsEditMode(false);
      
      // Update the local state with the updated game data
      setCurrentGame(updatedGame);
      
      // Refresh the games list
      await onGameUpdated();
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'Failed to save changes');
      console.error('Error updating game:', err);
    } finally {
      setIsSaving(false);
    }
  };
  
  // Handle copying a game
  const handleCopyGame = async () => {
    if (!copyName.trim()) {
      setCopyError('Game name is required');
      return;
    }
    
    try {
      setIsCopying(true);
      setCopyError(null);
      
      // Call the API to copy the game
      const copiedGame = await copyGame(game.id, copyName.trim());
      
      // Close the modal
      setIsCopyModalOpen(false);
      setCopyName('');
      
      // Refresh the games list
      await onGameUpdated();
      
      // Open the newly copied game
      // For now, we'll just update the current game with the copied one
      setCurrentGame(copiedGame);
    } catch (err) {
      setCopyError(err instanceof Error ? err.message : 'Failed to copy game');
      console.error('Error copying game:', err);
    } finally {
      setIsCopying(false);
    }
  };
  
  // Handle deleting a game
  const handleDeleteGame = async () => {
    try {
      setIsDeleting(true);
      setDeleteError(null);
      
      // Call the API to delete the game
      await deleteGame(currentGame.id);
      
      // Close the modal
      setIsDeleteModalOpen(false);
      
      // Refresh the games list
      await onGameUpdated();
      
      // Close the game details view
      onClose();
    } catch (err) {
      setDeleteError(err instanceof Error ? err.message : 'Failed to delete game');
      console.error('Error deleting game:', err);
    } finally {
      setIsDeleting(false);
    }
  };
  
  // Handle exporting a game
  const handleExport = async () => {
    try {
      // Get the game data from the API
      const response = await axios.get(`/api/v1/games/${currentGame.id}`);
      
      // Create a Blob from the JSON data
      const blob = new Blob([JSON.stringify(response.data, null, 2)], {
        type: 'application/json'
      });
      
      // Create a URL for the Blob
      const url = window.URL.createObjectURL(blob);
      
      // Create a temporary link element
      const link = document.createElement('a');
      link.href = url;
      link.download = `${currentGame.id}.json`;
      
      // Append to the document, click it, then remove it
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      // Release the URL object
      window.URL.revokeObjectURL(url);
    } catch (err) {
      console.error('Error exporting game:', err);
      alert('Failed to export game: ' + (err instanceof Error ? err.message : 'Unknown error'));
    }
  };
  
  return (
    <div className="game-details-container">
      <div className="game-details-header">
        <div className="game-buttons">
          <button className="delete-button" onClick={() => setIsDeleteModalOpen(true)}>Delete</button>
          <button className="copy-button" onClick={() => setIsCopyModalOpen(true)}>Copy</button>
          {isEditMode ? (
            <>
              <button 
                className="cancel-button" 
                onClick={handleCancelEdit}
              >
                Cancel
              </button>
              <button 
                className="save-button" 
                onClick={handleSaveEdit}
                disabled={isSaving}
              >
                {isSaving ? 'Saving...' : 'Save'}
              </button>
            </>
          ) : (
            <button className="edit-button" onClick={handleEditClick}>Edit</button>
          )}
          <button className="export-button" onClick={handleExport}>Export</button>
          <button className="close-button" onClick={onClose}>Close</button>
        </div>
      </div>

      <div className="game-details-content">
        <div className="game-details-title">
          <h2>{currentGame.name}</h2>
        </div>
        <div className="game-info-layout">
          <div className="game-icon-container">
            <img 
              src={iconSrc} 
              alt={`${currentGame.name} icon`} 
              className="game-icon" 
              width="128" 
              height="128"
            />
          </div>
          <div className="game-info-fields">
            <div className="game-field-container">
              <textarea 
                id="description"
                className="description-textarea"
                placeholder="Enter description"
                readOnly={!isEditMode}
                value={isEditMode ? editedDescription : (currentGame.description || '')}
                onChange={(e) => setEditedDescription(e.target.value)}
              />
            </div>
            <div className="game-field-container">
              <input
                type="text"
                id="tags"
                value={isEditMode ? editedTags : (currentGame.tags?.join(', ') || '')}
                onChange={(e) => setEditedTags(e.target.value)}
                placeholder="Add tags"
                className="tags-input"
                readOnly
              />
            </div>
          </div>
        </div>

        <div className="client-container">
          {currentGame.id && <GameIframe gameId={currentGame.id} />}
        </div>

        <div className="ai-field-container">
          <div className="response-container">
            <textarea 
              id="response"
              className="response-textarea"
              readOnly
              value={currentGame.response || ''}
            />
          </div>
          <div className="prompt-container">
            <textarea
              id="prompt"
              value={currentGame.ai_data?.prompt || ''}
              placeholder="Enter prompt"
              className="prompt-textarea"
            />
            <button className="prompt-button" onClick={onClose}>Prompt</button>
            <button className={currentGame.previous_id ? "undo-button" : "undo-button-disabled"} onClick={onClose}>Undo</button>
          </div>
        </div>

        <div className="game-additional-fields">
          <div className="game-field-container">
            <label htmlFor='id'>ID: </label>
            <input
              type="text"
              id="id"
              value={currentGame.id}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="game-field-container">
            <label htmlFor='name'>Name: </label>
            <input
              type="text"
              id="name"
              value={isEditMode ? editedName : currentGame.name}
              onChange={(e) => setEditedName(e.target.value)}
              placeholder="Enter name"
              className="name-input"
              readOnly={!isEditMode}
            />
          </div>
          <div className="game-field-container">
            <label htmlFor='status'>Status: </label>
            <input
              type="text"
              id="status"
              value={currentGame.status || ''}
              className="readonly-input"
              readOnly
            />
          </div>

        {currentGame.status_data && (
          <div className="status-data-container">
            <h3>Status Data:</h3>
            <pre className="status-data-json">
              {JSON.stringify(currentGame.status_data, null, 2)}
            </pre>
          </div>
        )}

          <div className="game-field-container">
            <label htmlFor='source'>Source: </label>
            <input
              type="text"
              id="source"
              value={currentGame.source || ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="game-field-container">
            <label htmlFor='created_at'>Created At: </label>
            <input
              type="text"
              id="created_at"
              value={currentGame.created_at ? new Date(currentGame.created_at * 1000).toLocaleString() : ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="game-field-container">
            <label htmlFor='created_by'>Created By: </label>
            <input
              type="text"
              id="created_by"
              value={currentGame.created_by || ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="game-field-container">
            <label htmlFor='updated_at'>Updated At: </label>
            <input
              type="text"
              id="updated_at"
              value={currentGame.updated_at ? new Date(currentGame.updated_at * 1000).toLocaleString() : ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="game-field-container">
            <label htmlFor='updated_by'>Updated By: </label>
            <input
              type="text"
              id="updated_by"
              value={currentGame.updated_by || ''}
              className="readonly-input"
              readOnly
            />
          </div>
        </div>
      </div>
      
      {saveError && <div className="modal-error">{saveError}</div>}
      
      {/* Copy Game Modal */}
      <Modal
        isOpen={isCopyModalOpen}
        onClose={() => {
          setIsCopyModalOpen(false);
          setCopyName('');
          setCopyError(null);
        }}
        title="Copy Game"
        actions={
          <>
            <button 
              className="cancel-button" 
              onClick={() => {
                setIsCopyModalOpen(false);
                setCopyName('');
                setCopyError(null);
              }}
            >
              Cancel
            </button>
            <button 
              className="action-button" 
              onClick={handleCopyGame}
              disabled={isCopying}
            >
              {isCopying ? 'Copying...' : 'Copy'}
            </button>
          </>
        }
      >
        <div className="modal-form-group">
          <label htmlFor="copy-game-name">New Game Name</label>
          <input
            id="copy-game-name"
            type="text"
            value={copyName}
            onChange={(e) => setCopyName(e.target.value)}
            placeholder="Enter a name for the copy"
            autoFocus
          />
        </div>
        {copyError && <div className="modal-error">{copyError}</div>}
      </Modal>
      
      {/* Delete Game Modal */}
      <Modal
        isOpen={isDeleteModalOpen}
        onClose={() => {
          setIsDeleteModalOpen(false);
          setDeleteError(null);
        }}
        title="Delete Game"
        actions={
          <>
            <button 
              className="cancel-button" 
              onClick={() => {
                setIsDeleteModalOpen(false);
                setDeleteError(null);
              }}
            >
              Cancel
            </button>
            <button 
              className="delete-action-button" 
              onClick={handleDeleteGame}
              disabled={isDeleting}
            >
              {isDeleting ? 'Deleting...' : 'Delete'}
            </button>
          </>
        }
      >
        <p>Are you sure you want to delete <strong>{currentGame.name}</strong>?</p>
        <p>This action cannot be undone.</p>
        {deleteError && <div className="modal-error">{deleteError}</div>}
      </Modal>
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
