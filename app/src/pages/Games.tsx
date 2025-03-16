import React from 'react';
import HelloWorld from '../components/HelloWorld';
import avatarLogo from '../assets/avatar.png';

const Games: React.FC = () => {
  return (
    <div className="games-container">
      <img src={avatarLogo} alt="Logo" className="logo" />
      <h1>Games</h1>
      <p>Browse and manage your game2d projects</p>
      
      <div className="components-container">
        <HelloWorld />
        <HelloWorld title="Custom Component Title" />
      </div>
    </div>
  );
};

export default Games;
