import React from 'react';
import { Link } from 'react-router-dom';

const Welcome: React.FC = () => {
  return (
    <div className="welcome-container">
      <div className="welcome-content">
        <h1>Welcome to game2d.ai</h1>
        
        <p>A platform for 2D game development with generative A.I.</p>
        
        <p className="welcome-item">game2d.ai is an open-source framework for
        game development that combines a 2D game engine with Lua scripting and
        a WebAssembly client. Built with a declarative object schema,
        representable in JSON, it is designed for experimenting with game code
        and assets created with generative AI.</p>
        
        <p>Please sign in to access the application.</p>
        
        <Link to="/login" className="welcome-button">
          Sign In
        </Link>
      </div>
    </div>
  );
};

export default Welcome;
