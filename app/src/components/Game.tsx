import React, { useEffect, useRef, useState, ChangeEvent, useImperativeHandle } from 'react';
import { Game as GameType, copyGame, deleteGame, updateGame, fetchGame } from '../services/gameService';
import avatarImage from '../assets/avatar.png';
import Modal from './Modal';
import axios from 'axios';
import { useAuth } from '../contexts/AuthContext';

// Helper function to check if a user has the required scope
const hasScope = (scopes: string | undefined, requiredScope: string): boolean => {
  if (!scopes) return false;
  return scopes.split(' ').includes(requiredScope) || scopes.split(' ').includes('superuser');
};

interface GameProps {
  game: GameType;
  onClose: () => void;
  onGameUpdated: () => Promise<void>; // Function to refresh games list
}

const Game: React.FC<GameProps> = ({ game, onClose, onGameUpdated }) => {
  const { user: authUser } = useAuth();

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

  // Tag modal states
  const [isAddTagModalOpen, setIsAddTagModalOpen] = useState(false);
  const [isDeleteTagModalOpen, setIsDeleteTagModalOpen] = useState(false);
  const [newTag, setNewTag] = useState('');
  const [selectedTag, setSelectedTag] = useState('');
  const [tagError, setTagError] = useState<string | null>(null);
  const [isTagProcessing, setIsTagProcessing] = useState(false);

  // Visibility change modal states
  const [isVisibilityModalOpen, setIsVisibilityModalOpen] = useState(false);
  const [newVisibilityValue, setNewVisibilityValue] = useState(false);
  const [visibilityError, setVisibilityError] = useState<string | null>(null);
  const [isVisibilityProcessing, setIsVisibilityProcessing] = useState(false);

  const clientIframeRef = useRef<HTMLIFrameElement | null>(null);
  const gameIframeRef = useRef<{ refreshIframe: () => void } | null>(null);

  // State to keep track of the current game data
  const [currentGame, setCurrentGame] = useState<GameType>(game);

  // State for prompt and response
  const [promptText, setPromptText] = useState('');
  const [responseText, setResponseText] = useState(game.ai_data?.response || '');
  const responseTextAreaRef = useRef<HTMLTextAreaElement>(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const [isUndo, setIsUndo] = useState(true); // true for Undo, false for Redo
  const [isPublic, setIsPublic] = useState(game.public || false);

  // Update polling state
  const [pollingInterval, setPollingInterval] = useState<ReturnType<typeof setInterval> | null>(null);
  const isUpdatingStatus = currentGame.status === "updating";

  // Check if user has write permission and belongs to the game's account
  const hasWritePermission = authUser &&
    (hasScope(authUser.scopes, 'games:write') &&
      (authUser.accountId === currentGame.account_id || hasScope(authUser.scopes, 'superuser')));

  // Update form fields when game prop changes
  useEffect(() => {
    setCurrentGame(game);
    setEditedName(game.name);
    setEditedDescription(game.description || '');
    setEditedTags(game.tags?.join(', ') || '');
    setResponseText(game.ai_data?.response || '');
    setIsUndo(true); // Reset to Undo when game changes
    setIsPublic(game.public || false);
    if (clientIframeRef.current) {
      clientIframeRef.current.src = `/client?game_id=${game.id}&game_name=${game.name}&api_url=${window.location.origin}/api/v1&api_token=${localStorage.getItem('token')}`;
    }
  }, [game]);

  // Set up polling for game status updates when status is "updating"
  useEffect(() => {
    // Clear any existing polling interval
    if (pollingInterval) {
      clearInterval(pollingInterval);
      setPollingInterval(null);
    }

    // If game status is "updating", set up a polling interval
    if (currentGame.status === "updating") {
      const interval = setInterval(async () => {
        try {
          // Fetch the latest game data
          const updatedGame = await fetchGame(currentGame.id);
          setCurrentGame(updatedGame);

          // If status is no longer "updating", clear the interval
          if (updatedGame.status !== "updating") {
            clearInterval(interval);
            setPollingInterval(null);
            setResponseText(updatedGame.ai_data?.response || '');
            // Refresh the games table
            await onGameUpdated();
          }
        } catch (err) {
          console.error("Error polling game status:", err);
          clearInterval(interval);
          setPollingInterval(null);
        }
      }, 1000); // Poll every second

      setPollingInterval(interval);

      // Cleanup function to clear interval when component unmounts or game changes
      return () => {
        clearInterval(interval);
      };
    }
  }, [currentGame.status, currentGame.id, onGameUpdated]);

  useEffect(() => {
    if (responseTextAreaRef.current) {
      responseTextAreaRef.current.scrollTop = responseTextAreaRef.current.scrollHeight;
    }
  }, [responseText]);

  // Handle public checkbox change
  const handlePublicChange = (e: ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.checked;

    // Set visibility value for the modal
    setNewVisibilityValue(newValue);

    // Temporarily update checkbox visual state
    setIsPublic(newValue);

    // Reset error state
    setVisibilityError(null);

    // Open confirmation modal
    setIsVisibilityModalOpen(true);
  };

  // Handle confirming the visibility change
  const handleConfirmVisibilityChange = async () => {
    try {
      setIsVisibilityProcessing(true);
      setVisibilityError(null);

      // Prepare the update payload
      const updates: Partial<GameType> = {
        public: newVisibilityValue
      };

      // Call the API to update the game
      const updatedGame = await updateGame(currentGame.id, updates);

      // Update the local state with the updated game data
      setCurrentGame(updatedGame);

      // Close modal
      setIsVisibilityModalOpen(false);

      // Refresh the games list
      await onGameUpdated();
    } catch (err) {
      // Show error in modal
      setVisibilityError(err instanceof Error ? err.message : 'Failed to update game visibility');
      console.error('Error updating public status:', err);

      // Revert to previous value if there was an error
      setIsPublic(!newVisibilityValue);
    } finally {
      setIsVisibilityProcessing(false);
    }
  };

  // Handle canceling the visibility change
  const handleCancelVisibilityChange = () => {
    // Revert to original value
    setIsPublic(currentGame.public || false);

    // Close the modal
    setIsVisibilityModalOpen(false);
    setVisibilityError(null);
  };

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

  // Handle the Prompt button click
  const handlePromptClick = async () => {
    if (!promptText.trim()) {
      return; // Don't send empty prompts
    }

    try {
      setIsProcessing(true);

      // Make the POST request to the prompt API
      const response = await axios.post('/api/v1/games/prompt', {
        prompt: promptText,
        game_id: currentGame.id
      });

      // Clear the prompt textarea
      setPromptText('');

      // Reset to Undo state
      setIsUndo(true);

      // Fetch the updated game to refresh the component
      const updatedGame = await fetchGame(response.data.game_id);
      setCurrentGame(updatedGame);
      setResponseText(updatedGame.ai_data?.response || '');
      if (responseTextAreaRef.current) {
        responseTextAreaRef.current.scrollTop = responseTextAreaRef.current.scrollHeight;
      }

      // Refresh the games table
      await onGameUpdated();
    } catch (err) {
      // Append error to response textarea
      const errorMessage = err instanceof Error ? err.message : 'Error processing prompt';
      setResponseText((prev: string) => {
        const separator = prev ? '\n\n' : '';
        return `${prev}${separator}ERROR: ${errorMessage}`;
      });
      console.error('Error processing prompt:', err);
    } finally {
      setIsProcessing(false);
    }
  };

  // Handle the Undo/Redo button click
  const handleUndoRedoClick = async () => {
    try {
      setIsProcessing(true);

      // Make the POST request to the undo API
      const response = await axios.post('/api/v1/games/undo', {
        game_id: currentGame.id
      });

      // If successful, refresh with the new game ID
      if (response.data && response.data.game_id) {
        // Toggle between Undo and Redo
        setIsUndo(!isUndo);

        // Fetch the updated game
        const updatedGame = await fetchGame(response.data.game_id);
        setCurrentGame(updatedGame);
        setResponseText(updatedGame.ai_data?.response || '');
        if (responseTextAreaRef.current) {
          responseTextAreaRef.current.scrollTop = responseTextAreaRef.current.scrollHeight;
        }

        // Refresh the games table
        await onGameUpdated();
      }
    } catch (err) {
      // Append error to response textarea
      const errorMessage = err instanceof Error ? err.message : 'unable to process undo/redo';
      setResponseText((prev: string) => {
        const separator = prev ? '\n\n' : '';
        return `${prev}${separator}ERROR: ${errorMessage}`;
      });
      console.error('Error processing undo/redo:', err);
    } finally {
      setIsProcessing(false);
    }
  };

  // Handle closing the game and refreshing the table
  const handleCloseAndRefresh = async () => {
    // Refresh the games list
    await onGameUpdated();

    // Close the game details view
    onClose();
  };

  // Handle refreshing a game
  const handleRefresh = async () => {
    try {
      // Fetch updated game data
      const updatedGame = await fetchGame(currentGame.id);
      setCurrentGame(updatedGame);
      
      // Refresh the iframe content
      if (gameIframeRef.current) {
        gameIframeRef.current.refreshIframe();
      }
      
      // Refresh the games list
      await onGameUpdated();
    } catch (err) {
      console.error("Error refreshing game:", err);
    }
  }

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
          {hasWritePermission && (
            <>
              <button
                className="delete-button"
                onClick={() => setIsDeleteModalOpen(true)}
                disabled={isUpdatingStatus}
              >
                Delete
              </button>
              <button
                className="copy-button"
                onClick={() => setIsCopyModalOpen(true)}
                disabled={isUpdatingStatus}
              >
                Copy
              </button>
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
                <button
                  className="edit-button"
                  onClick={handleEditClick}
                  disabled={isUpdatingStatus}
                >
                  Edit
                </button>
              )}
            </>
          )}
          <button
            className="export-button"
            onClick={handleExport}
          >
            Export
          </button>
          <button
            className="refresh-button"
            onClick={handleRefresh}
          >
            Refresh
          </button>
          <button className="close-button" onClick={handleCloseAndRefresh}>Close</button>
        </div>
      </div>

      {saveError && <div className="modal-error">{saveError}</div>}

      <div className="game-details-content">
        <div className="game-details-title">
          <h2>{currentGame.name}</h2>
          <div className="game-field-container">
            <label htmlFor="public-checkbox" className="wide">Public:</label>
            <input
              id="public-checkbox"
              type="checkbox"
              checked={isPublic}
              onChange={handlePublicChange}
              disabled={isProcessing || !hasWritePermission}
              className="wide"
              readOnly={!hasWritePermission}
            />
          </div>
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
          </div>
        </div>
        <div className="game-info-tags">
          {hasWritePermission ? (
            <button
              className="add-tag-button"
              onClick={() => {
                setNewTag('');
                setTagError(null);
                setIsAddTagModalOpen(true);
              }}
              disabled={isEditMode || isUpdatingStatus}
            >
              Add Tag
            </button>
          ) : (
            <span className="tags-label">Tags:</span>
          )}

          {currentGame.tags && currentGame.tags.length > 0 ? (
            currentGame.tags.map((tag, index) => (
              <span
                key={index}
                className={hasScope(authUser?.scopes, 'user:write') && !isEditMode && !isUpdatingStatus ? "tag-item clickable" : "tag-item"}
                onClick={() => {
                  if (hasScope(authUser?.scopes, 'user:write') && !isEditMode && !isUpdatingStatus) {
                    setSelectedTag(tag);
                    setTagError(null);
                    setIsDeleteTagModalOpen(true);
                  }
                }}
              >
                {tag}
              </span>
            ))
          ) : (
            <span style={{ color: '#888', fontSize: '0.9rem' }}>No tags</span>
          )}
        </div>

        <div className="client-container">
          {currentGame.id && (
            <GameIframe 
              ref={gameIframeRef}
              gameId={currentGame.id} 
              gameName={currentGame.name} 
            />
          )}
        </div>

        {hasWritePermission && (
          <div className="ai-field-container">
            <div className="response-container">
              <textarea
                ref={responseTextAreaRef}
                id="response"
                className="response-textarea"
                readOnly
                onLoad={() => { setResponseText(currentGame.ai_data?.response || '') }}
                onChange={(e) => setResponseText(e.target.value)}
                value={responseText}
              />
            </div>
            <div className="prompt-container">
              <textarea
                id="prompt"
                value={promptText}
                onChange={(e) => setPromptText(e.target.value)}
                placeholder="Enter prompt"
                className="prompt-textarea"
                disabled={isEditMode || isProcessing || isUpdatingStatus}
              />
              <button
                className={isUpdatingStatus ? "prompt-button updating-button" : "prompt-button"}
                onClick={handlePromptClick}
                disabled={isEditMode || isProcessing || isUpdatingStatus}
              >
                {isUpdatingStatus ? 'Updating...' : isProcessing ? 'Processing...' : 'Prompt'}
              </button>
              <button
                className="undo-button"
                onClick={handleUndoRedoClick}
                disabled={isEditMode || isProcessing || (!currentGame.previous_id)}
              >
                {isUpdatingStatus ? 'Cancel' : isUndo ? 'Undo' : 'Redo'}
              </button>
            </div>
          </div>
        )}

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

      {/* Add Tag Modal */}
      <Modal
        isOpen={isAddTagModalOpen}
        onClose={() => {
          setIsAddTagModalOpen(false);
          setTagError(null);
        }}
        title="Add Tag"
        actions={
          <>
            <button
              className="cancel-button"
              onClick={() => {
                setIsAddTagModalOpen(false);
                setTagError(null);
              }}
            >
              Cancel
            </button>
            <button
              className="action-button"
              onClick={async () => {
                if (!newTag.trim()) {
                  setTagError('Tag cannot be empty');
                  return;
                }

                try {
                  setIsTagProcessing(true);
                  setTagError(null);

                  // POST request to add tag
                  await axios.post(`/api/v1/games/${currentGame.id}/tags`, [newTag.trim()]);

                  // Close modal
                  setIsAddTagModalOpen(false);
                  setNewTag('');

                  // Refresh game to show new tag
                  const updatedGame = await fetchGame(currentGame.id);
                  setCurrentGame(updatedGame);
                  await onGameUpdated();
                } catch (err) {
                  setTagError(err instanceof Error ? err.message : 'Failed to add tag');
                  console.error('Error adding tag:', err);
                } finally {
                  setIsTagProcessing(false);
                }
              }}
              disabled={isTagProcessing}
            >
              {isTagProcessing ? 'Adding...' : 'Add'}
            </button>
          </>
        }
      >
        <div className="modal-form-group">
          <label htmlFor="tag-input">New Tag</label>
          <input
            id="tag-input"
            type="text"
            value={newTag}
            onChange={(e) => setNewTag(e.target.value)}
            placeholder="Enter tag"
            autoFocus
          />
        </div>
        {tagError && <div className="modal-error">{tagError}</div>}
      </Modal>

      {/* Delete Tag Modal */}
      <Modal
        isOpen={isDeleteTagModalOpen}
        onClose={() => {
          setIsDeleteTagModalOpen(false);
          setTagError(null);
        }}
        title="Delete Tag"
        actions={
          <>
            <button
              className="cancel-button"
              onClick={() => {
                setIsDeleteTagModalOpen(false);
                setTagError(null);
              }}
            >
              Cancel
            </button>
            <button
              className="delete-action-button"
              onClick={async () => {
                try {
                  setIsTagProcessing(true);
                  setTagError(null);

                  // DELETE request to remove tag
                  await axios.delete(`/api/v1/games/${currentGame.id}/tags`, {
                    data: [selectedTag]
                  });

                  // Close modal
                  setIsDeleteTagModalOpen(false);
                  setSelectedTag('');

                  // Refresh game to update tags
                  const updatedGame = await fetchGame(currentGame.id);
                  setCurrentGame(updatedGame);
                  await onGameUpdated();
                } catch (err) {
                  setTagError(err instanceof Error ? err.message : 'Failed to delete tag');
                  console.error('Error deleting tag:', err);
                } finally {
                  setIsTagProcessing(false);
                }
              }}
              disabled={isTagProcessing}
            >
              {isTagProcessing ? 'Deleting...' : 'Yes'}
            </button>
          </>
        }
      >
        <p>Are you sure you want to delete the tag <strong>{selectedTag}</strong>?</p>
        {tagError && <div className="modal-error">{tagError}</div>}
      </Modal>

      {/* Visibility Change Modal */}
      <Modal
        isOpen={isVisibilityModalOpen}
        onClose={handleCancelVisibilityChange}
        title="Change Visibility"
        actions={
          <>
            <button
              className="cancel-button"
              onClick={handleCancelVisibilityChange}
            >
              Cancel
            </button>
            <button
              className="action-button"
              onClick={handleConfirmVisibilityChange}
              disabled={isVisibilityProcessing}
            >
              {isVisibilityProcessing ? 'Processing...' : 'Confirm'}
            </button>
          </>
        }
      >
        <p>
          Are you sure you want to make the game <strong>{newVisibilityValue ? 'public' : 'private'}</strong>?
        </p>
        {visibilityError && <div className="modal-error">{visibilityError}</div>}
      </Modal>
    </div>
  );
};

// Component for the game iframe that handles token retrieval and refresh logic
const GameIframe = React.forwardRef<
  { refreshIframe: () => void },
  { gameId: string; gameName: string }
>(({ gameId, gameName }, ref) => {
  const [apiToken, setApiToken] = useState<string>('');
  const iframeRef = useRef<HTMLIFrameElement>(null);

  // Get API token from localStorage on component mount and when gameId changes
  useEffect(() => {
    const token = localStorage.getItem('token') || '';
    setApiToken(token);
  }, [gameId]);

  // Expose refreshIframe method through ref
  useImperativeHandle(ref, () => ({
    refreshIframe: () => {
      if (iframeRef.current) {
        // Update the iframe src to refresh it
        const currentSrc = iframeRef.current.src;
        iframeRef.current.src = '';
        setTimeout(() => {
          if (iframeRef.current) {
            iframeRef.current.src = currentSrc;
          }
        }, 50);
      }
    }
  }));

  // Build the iframe src URL with query parameters
  const iframeSrc = apiToken ?
    `/client?game_id=${gameId}&game_name=${gameName}&api_url=${window.location.origin}/api/v1&api_token=${apiToken}` : '';

  // Only render the iframe when we have an API token
  return apiToken ? (
    <iframe
      ref={iframeRef}
      src={iframeSrc}
      title="Game Client"
      width="640px"
      height="480px"
      frameBorder="0"
      allowFullScreen
    />
  ) : null;
});

export default Game;
