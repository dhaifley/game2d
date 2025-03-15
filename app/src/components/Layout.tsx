import React from 'react';
import { Outlet } from 'react-router-dom';
import TitleBar from './TitleBar';

const Layout: React.FC = () => {
  return (
    <div className="app-layout">
      <TitleBar />
      <div className="app-content">
        <Outlet />
      </div>
    </div>
  );
};

export default Layout;
