import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const TitleBar: React.FC = () => {
  const { isAuthenticated, user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const currentPath = location.pathname;

  const handleSignInClick = () => {
    navigate('/login');
  };

  const handleSignOutClick = () => {
    logout();
    navigate('/welcome');
  };
  
  const handleHelpClick = () => {
    navigate('/help');
  };
  
  const handleGamesClick = () => {
    navigate('/');
  };

  return (
    <div className="title-bar">
      <div className="title">game2d.ai</div>
      <div className="title-buttons">
        <button className={currentPath === '/' ? "games-button-sel" : "games-button"} onClick={handleGamesClick}>
          Games
        </button>
        <button className={currentPath === '/help' ? "help-button-sel" : "help-button"} onClick={handleHelpClick}>
          Help
        </button>
        {isAuthenticated && user ? (
          <button className="user-button">
            {user.id}
          </button>
        ) : (
          <button className={currentPath === '/login' ? "sign-in-button-sel" : "sign-in-button"} onClick={handleSignInClick}>
            Sign In
          </button>
        )}
        {isAuthenticated && (
          <button className="sign-out-button" onClick={handleSignOutClick}>
            Sign Out
          </button>
        )}
      </div>
    </div>
  );
};

export default TitleBar;
