import { useState, useEffect, forwardRef, useImperativeHandle } from 'react';
import avatarLogo from '../assets/avatar.png';
import { Game, fetchGames, createGame } from '../services/gameService';
import Modal from './Modal';

interface SortConfig {
  key: string;
  direction: 'asc' | 'desc' | null;
}

interface GamesTableProps {
  onSelectGame: (game: Game) => void;
}

// Define the ref handle type
export interface GamesTableHandle {
  loadGames: () => Promise<void>;
}

const GamesTable = forwardRef<GamesTableHandle, GamesTableProps>(({ onSelectGame }, ref) => {
  const [games, setGames] = useState<Game[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [pageSize, setPageSize] = useState(10);
  const [pageSkip, setPageSkip] = useState(0);
  const [sortConfig, setSortConfig] = useState<SortConfig>({ key: 'name', direction: 'asc' });
  const [totalGames, setTotalGames] = useState(0);
  
  // New Game Modal State
  const [isNewGameModalOpen, setIsNewGameModalOpen] = useState(false);
  const [newGameName, setNewGameName] = useState('');
  const [newGameError, setNewGameError] = useState<string | null>(null);
  const [creatingGame, setCreatingGame] = useState(false);

  // Expose the loadGames method to parent components via ref
  useImperativeHandle(ref, () => ({
    loadGames
  }));
  
  const loadGames = async () => {
    try {
      setLoading(true);
      setError(null);

      const result = await fetchGames(
        searchQuery,
        pageSize,
        pageSkip,
        sortConfig
      );
      
      setGames(result.games);
      setTotalGames(result.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
      console.error('Error fetching games:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadGames();
  }, [pageSize, pageSkip, searchQuery, sortConfig]);

  const handleSort = (key: string) => {
    setSortConfig(prevConfig => {
      if (prevConfig.key === key) {
        // Cycle through: asc -> desc -> null -> asc
        const nextDirection = 
          prevConfig.direction === 'asc' ? 'desc' : 
          prevConfig.direction === 'desc' ? null : 'asc';
        
        return {
          key,
          direction: nextDirection
        };
      }
      
      // New column selected - start with ascending
      return { key, direction: 'asc' };
    });
  };

  const getSortIndicator = (key: string) => {
    if (sortConfig.key !== key) return null;
    
    if (sortConfig.direction === 'asc') return '↑';
    if (sortConfig.direction === 'desc') return '↓';
    return null;
  };

  const handleNextPage = () => {
    setPageSkip(prev => prev + pageSize);
  };

  const handlePrevPage = () => {
    setPageSkip(prev => Math.max(0, prev - pageSize));
  };

  const renderIcon = (game: Game) => {
    if (game.icon) {
      return (
        <div className="game-icon-container">
          <img 
            src={`data:image/svg+xml;base64,${game.icon}`} 
            alt={`${game.name} icon`} 
            className="game-icon" 
          />
        </div>
      );
    } else {
      return (
        <div className="game-icon-container">
          <img 
            src={avatarLogo} 
            alt="Default icon" 
            className="game-icon" 
            style={{ width: '32px', height: '32px' }}
          />
        </div>
      );
    }
  };

  const formatDate = (timestamp?: number) => {
    if (!timestamp) return '';
    return new Date(timestamp * 1000).toLocaleString();
  };

  return (
    <div className="games-table-container">
      <div className="games-table-controls">
        <div className="search-container">
          <input
            type="text"
            placeholder="Search games..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="search-input"
          />
        </div>
        <div className="page-size-selector">
          <label htmlFor="page-size">Page size:</label>
          <select 
            id="page-size" 
            value={pageSize} 
            onChange={(e) => setPageSize(Number(e.target.value))}
          >
            <option value={5}>5</option>
            <option value={10}>10</option>
            <option value={25}>25</option>
            <option value={50}>50</option>
          </select>
        </div>
        <div className="import-game-button">
          <button className="import-button" onClick={() => {}}>Import</button>
        </div>
        <div className="new-game-button">
          <button className="new-button" onClick={() => setIsNewGameModalOpen(true)}>New Game</button>
        </div>
      </div>

      {loading && <div className="loading-indicator">Loading games...</div>}
      {error && <div className="error-message">Error: {error}</div>}
      
      {!loading && !error && games.length === 0 && (
        <div className="no-results">No games found</div>
      )}

      {!loading && !error && games.length > 0 && (
        <>
          <div className="games-table-wrapper">
            <table className="games-table">
              <thead>
                <tr>
                  <th></th>
                  <th onClick={() => handleSort('name')}>
                    Name {getSortIndicator('name')}
                  </th>
                  <th onClick={() => handleSort('version')}>
                    Version {getSortIndicator('version')}
                  </th>
                  <th onClick={() => handleSort('status')}>
                    Status {getSortIndicator('status')}
                  </th>
                  <th onClick={() => handleSort('source')}>
                    Source {getSortIndicator('source')}
                  </th>
                  <th onClick={() => handleSort('updated_at')}>
                    Updated At {getSortIndicator('updated_at')}
                  </th>
                </tr>
              </thead>
              <tbody>
                {games.map(game => (
                  <tr 
                    key={game.id} 
                    onClick={() => onSelectGame(game)}
                    className="game-row"
                  >
                    <td>{renderIcon(game)}</td>
                    <td>{game.name}</td>
                    <td>{game.version || ''}</td>
                    <td>{game.status || ''}</td>
                    <td>{game.source || ''}</td>
                    <td>{formatDate(game.updated_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="pagination-controls">
            <button 
              onClick={handlePrevPage} 
              disabled={pageSkip === 0}
              className="pagination-button"
            >
              Previous
            </button>
            <span className="pagination-info">
              Showing {pageSkip + 1} - {pageSkip + games.length}
              {totalGames > 0 ? ` of ${totalGames}` : ''}
            </span>
            <button 
              onClick={handleNextPage} 
              disabled={pageSkip + games.length >= totalGames}
              className="pagination-button"
            >
              Next
            </button>
          </div>
        </>
      )}

      {/* New Game Modal */}
      <Modal
        isOpen={isNewGameModalOpen}
        onClose={() => {
          setIsNewGameModalOpen(false);
          setNewGameName('');
          setNewGameError(null);
        }}
        title="Create New Game"
        actions={
          <>
            <button 
              className="cancel-button" 
              onClick={() => {
                setIsNewGameModalOpen(false);
                setNewGameName('');
                setNewGameError(null);
              }}
            >
              Cancel
            </button>
            <button 
              className="action-button" 
              onClick={async () => {
                if (!newGameName.trim()) {
                  setNewGameError('Game name is required');
                  return;
                }

                try {
                  setCreatingGame(true);
                  setNewGameError(null);
                  const newGame = await createGame(newGameName.trim());
                  
                  // Close the modal
                  setIsNewGameModalOpen(false);
                  setNewGameName('');
                  
                  // Refresh the games list
                  await loadGames();
                  
                  // Open the newly created game
                  onSelectGame(newGame);
                } catch (err) {
                  setNewGameError(err instanceof Error ? err.message : 'Failed to create game');
                  console.error('Error creating game:', err);
                } finally {
                  setCreatingGame(false);
                }
              }}
              disabled={creatingGame}
            >
              {creatingGame ? 'Creating...' : 'Create'}
            </button>
          </>
        }
      >
        <div className="modal-form-group">
          <label htmlFor="new-game-name">Game Name</label>
          <input
            id="new-game-name"
            type="text"
            value={newGameName}
            onChange={(e) => setNewGameName(e.target.value)}
            placeholder="Enter a name for your new game"
            autoFocus
          />
        </div>
        {newGameError && <div className="modal-error">{newGameError}</div>}
      </Modal>
    </div>
  );
});

export default GamesTable;
