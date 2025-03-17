import React, { useState, useEffect } from 'react';
import avatarLogo from '../assets/avatar.png';
import { Game, fetchGames } from '../services/gameService';

interface SortConfig {
  key: string;
  direction: 'asc' | 'desc' | null;
}

interface GamesTableProps {
  onSelectGame: (game: Game) => void;
}

const GamesTable: React.FC<GamesTableProps> = ({ onSelectGame }) => {
  const [games, setGames] = useState<Game[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [pageSize, setPageSize] = useState(10);
  const [pageSkip, setPageSkip] = useState(0);
  const [sortConfig, setSortConfig] = useState<SortConfig>({ key: 'name', direction: 'asc' });
  const [totalGames, setTotalGames] = useState(0);

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
    if (!timestamp) return 'N/A';
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
          <label htmlFor="page-size">Items per page:</label>
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
                  <th>Icon</th>
                  <th onClick={() => handleSort('name')}>
                    Name {getSortIndicator('name')}
                  </th>
                  <th onClick={() => handleSort('version')}>
                    Version {getSortIndicator('version')}
                  </th>
                  <th onClick={() => handleSort('status')}>
                    Status {getSortIndicator('status')}
                  </th>
                  <th onClick={() => handleSort('updated_at')}>
                    Updated At {getSortIndicator('updated_at')}
                  </th>
                  <th onClick={() => handleSort('updated_by')}>
                    Updated By {getSortIndicator('updated_by')}
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
                    <td>{game.version || 'N/A'}</td>
                    <td>{game.status || 'N/A'}</td>
                    <td>{formatDate(game.updated_at)}</td>
                    <td>{game.updated_by || 'N/A'}</td>
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
              disabled={games.length < pageSize}
              className="pagination-button"
            >
              Next
            </button>
          </div>
        </>
      )}
    </div>
  );
};

export default GamesTable;
