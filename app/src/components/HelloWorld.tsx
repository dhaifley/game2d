import React from 'react';
import avatarLogo from '../assets/avatar.png';

interface HelloWorldProps {
  title?: string;
}

const HelloWorld: React.FC<HelloWorldProps> = ({ title = 'Hello World' }) => {
  return (
    <div className="hello-world-component">
      <img src={avatarLogo} alt="Avatar Logo" className="avatar-logo" />
      <h2>{title}</h2>
      <p>This is a simple React component in the game2d-app</p>
    </div>
  );
};

export default HelloWorld;
