import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const TitleBar: React.FC = () => {
  const { isAuthenticated, user, logout } = useAuth();
  const navigate = useNavigate();

  const handleSignInClick = () => {
    navigate('/login');
  };

  const handleSignOutClick = () => {
    logout();
    navigate('/welcome');
  };

  return (
    <div className="title-bar">
      <div className="title">game2d.ai</div>
      <div className="auth-buttons">
        {isAuthenticated && user && (
          <button className="user-button">
            {user.id}
          </button>
        )}
        {isAuthenticated ? (
          <button className="sign-out-button" onClick={handleSignOutClick}>
            Sign Out
          </button>
        ) : (
          <button className="sign-in-button" onClick={handleSignInClick}>
            Sign In
          </button>
        )}
      </div>
    </div>
  );
};

export default TitleBar;
