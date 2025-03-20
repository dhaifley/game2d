import React from 'react';
import Account from '../components/Account';
import { useNavigate } from 'react-router-dom';

const AccountPage: React.FC = () => {
  const navigate = useNavigate();

  const handleClose = () => {
    navigate('/');
  };

  return (
    <div className="account-page">
      <Account onClose={handleClose} />
    </div>
  );
};

export default AccountPage;
