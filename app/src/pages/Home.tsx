import React from 'react';
import HelloWorld from '../components/HelloWorld';
import avatarLogo from '../assets/avatar.png';

const Home: React.FC = () => {
  return (
    <div className="home-container">
      <img src={avatarLogo} alt="Logo" className="logo" />
      <h1>Hello World!</h1>
      <p>Welcome to the game2d-app</p>
      
      <div className="components-container">
        <HelloWorld />
        <HelloWorld title="Custom Component Title" />
      </div>
    </div>
  );
};

export default Home;
