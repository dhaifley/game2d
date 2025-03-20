import React from 'react';
import User from '../components/User';
import { useNavigate } from 'react-router-dom';

const UserPage: React.FC = () => {
  const navigate = useNavigate();

  const handleClose = () => {
    navigate('/');
  };

  return (
    <div className="user-page">
      <User onClose={handleClose} />
    </div>
  );
};

export default UserPage;
