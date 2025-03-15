import React from 'react';
import { Link } from 'react-router-dom';

const Welcome: React.FC = () => {
  return (
    <div className="welcome-container">
      <div className="welcome-content">
        <h1>Welcome to game2d.ai</h1>
        <p>Your platform for 2D game development with generative A.I.</p>
        <p>Please sign in to access the application.</p>
        
        <Link to="/login" className="welcome-button">
          Sign In
        </Link>
      </div>
    </div>
  );
};

export default Welcome;
