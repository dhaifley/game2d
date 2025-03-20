import React, { useState, useRef, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const TitleBar: React.FC = () => {
  const { isAuthenticated, user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const currentPath = location.pathname;
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Close dropdown when clicking outside
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  const handleSignInClick = () => {
    navigate('/login');
  };

  const handleSignOutClick = () => {
    logout();
    navigate('/welcome');
    setIsDropdownOpen(false);
  };
  
  const handleHelpClick = () => {
    navigate('/help');
  };
  
  const handleGamesClick = () => {
    navigate('/');
  };

  const handleUserProfileClick = () => {
    navigate('/user');
    setIsDropdownOpen(false);
  };

  const handleAccountSettingsClick = () => {
    navigate('/account');
    setIsDropdownOpen(false);
  };

  const toggleDropdown = () => {
    setIsDropdownOpen(!isDropdownOpen);
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
          <div className="user-dropdown-container" ref={dropdownRef}>
            <button className="user-button" onClick={toggleDropdown}>
              {user.id}
            </button>
            {isDropdownOpen && (
              <div className="user-dropdown-menu">
                <button className="dropdown-item" onClick={handleUserProfileClick}>
                  User Profile
                </button>
                <button className="dropdown-item" onClick={handleAccountSettingsClick}>
                  Account Settings
                </button>
                <button className="dropdown-item" onClick={handleSignOutClick}>
                  Sign Out
                </button>
              </div>
            )}
          </div>
        ) : (
          <button className={currentPath === '/login' ? "sign-in-button-sel" : "sign-in-button"} onClick={handleSignInClick}>
            Sign In
          </button>
        )}
      </div>
    </div>
  );
};

export default TitleBar;
