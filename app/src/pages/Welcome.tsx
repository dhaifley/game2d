import React from 'react';
import { Link } from 'react-router-dom';

const Welcome: React.FC = () => {
  return (
    <div className="welcome-container">
      <div className="welcome-content">
        <h1>Welcome to game2d.ai</h1>
        
        <p>A platform for 2D game development with generative A.I.</p>
        
        <p className="welcome-item">game2d.ai is an open-source framework for
        2D game development that combines Go, the ebitengine game engine,
        Lua scripting, and a WebAssembly client. Built on a declarative object
        schema, representable in JSON or YAML, it's designed for experimenting
        with game code and assets created with generative AI.</p>
        
        <p>Please sign in to access the application.</p>
        
        <Link to="/login" className="welcome-button">
          Sign In
        </Link>
      </div>
    </div>
  );
};

export default Welcome;
