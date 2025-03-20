import React, { useEffect, useState } from 'react';
import { Account as AccountType, fetchAccount, updateAccount } from '../services/accountService';

interface AccountProps {
  onClose: () => void;
}

const Account: React.FC<AccountProps> = ({ onClose }) => {
  // Component state
  const [account, setAccount] = useState<AccountType | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Edit mode state
  const [isEditMode, setIsEditMode] = useState(false);
  const [editedName, setEditedName] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  // Load account data when component mounts
  useEffect(() => {
    const loadAccount = async () => {
      try {
        setLoading(true);
        setError(null);
        const accountData = await fetchAccount();
        setAccount(accountData);
        setEditedName(accountData.name || '');
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load account data');
        console.error('Error loading account:', err);
      } finally {
        setLoading(false);
      }
    };

    loadAccount();
  }, []);

  // Handle entering edit mode
  const handleEditClick = () => {
    setIsEditMode(true);
    setSaveError(null);
  };

  // Handle canceling edit mode (restore original values)
  const handleCancelEdit = () => {
    setIsEditMode(false);
    if (account) {
      setEditedName(account.name || '');
    }
    setSaveError(null);
  };

  // Handle saving edited account
  const handleSaveEdit = async () => {
    if (!account) return;
    
    try {
      setIsSaving(true);
      setSaveError(null);
      
      // Prepare the update payload
      const updates: Partial<AccountType> = {
        name: editedName.trim()
      };
      
      // Call the API to update the account
      const updatedAccount = await updateAccount(updates);
      
      // Exit edit mode
      setIsEditMode(false);
      
      // Update the local state with the updated account data
      setAccount(updatedAccount);
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'Failed to save changes');
      console.error('Error updating account:', err);
    } finally {
      setIsSaving(false);
    }
  };

  if (loading) {
    return <div className="loading-indicator">Loading account data...</div>;
  }

  if (error) {
    return <div className="error-message">Error: {error}</div>;
  }

  if (!account) {
    return <div className="error-message">No account data available</div>;
  }

  return (
    <div className="account-details-container">
      <div className="account-details-header">
        <div className="account-buttons">
          {isEditMode ? (
            <>
              <button 
                className="cancel-button" 
                onClick={handleCancelEdit}
              >
                Cancel
              </button>
              <button 
                className="save-button" 
                onClick={handleSaveEdit}
                disabled={isSaving}
              >
                {isSaving ? 'Saving...' : 'Save'}
              </button>
            </>
          ) : (
            <button className="edit-button" onClick={handleEditClick}>Edit</button>
          )}
          <button className="close-button" onClick={onClose}>Close</button>
        </div>
      </div>

      <div className="account-details-content">
        <div className="account-details-title">
          <h2>{account.name}</h2>
        </div>

        <div className="account-additional-fields">
          <div className="account-field-container">
            <label htmlFor='id'>ID:</label>
            <input
              type="text"
              id="id"
              value={account.id}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="account-field-container">
            <label htmlFor='name'>Name:</label>
            <input
              type="text"
              id="name"
              value={isEditMode ? editedName : (account.name || '')}
              onChange={(e) => setEditedName(e.target.value)}
              className={isEditMode ? "name-input" : "readonly-input"}
              readOnly={!isEditMode}
            />
          </div>
          <div className="account-field-container">
            <label htmlFor='status'>Status:</label>
            <input
              type="text"
              id="status"
              value={account.status || ''}
              className="readonly-input"
              readOnly
            />
          </div>
          
          {account.status_data && (
            <div className="account-data-container">
              <h3>Status Data:</h3>
              <pre className="account-data-json">
                {JSON.stringify(account.status_data, null, 2)}
              </pre>
            </div>
          )}
          
          <div className="account-field-container">
            <label htmlFor='repo_status'>Repo Status:</label>
            <input
              type="text"
              id="repo_status"
              value={account.repo_status || ''}
              className="readonly-input"
              readOnly
            />
          </div>
          
          {account.repo_status_data && (
            <div className="account-data-container">
              <h3>Repo Status Data:</h3>
              <pre className="account-data-json">
                {JSON.stringify(account.repo_status_data, null, 2)}
              </pre>
            </div>
          )}
          
          <div className="account-field-container">
            <label htmlFor='game_limit'>Game Limit:</label>
            <input
              type="text"
              id="game_limit"
              value={account.game_limit?.toString() || ''}
              className="readonly-input"
              readOnly
            />
          </div>
          
          {account.data && (
            <div className="account-data-container">
              <h3>Account Data:</h3>
              <pre className="account-data-json">
                {JSON.stringify(account.data, null, 2)}
              </pre>
            </div>
          )}

          <div className="account-field-container">
            <label htmlFor='created_at'>Created At:</label>
            <input
              type="text"
              id="created_at"
              value={account.created_at ? new Date(account.created_at * 1000).toLocaleString() : ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="account-field-container">
            <label htmlFor='updated_at'>Updated At:</label>
            <input
              type="text"
              id="updated_at"
              value={account.updated_at ? new Date(account.updated_at * 1000).toLocaleString() : ''}
              className="readonly-input"
              readOnly
            />
          </div>
        </div>
      </div>
      
      {saveError && <div className="modal-error">{saveError}</div>}
    </div>
  );
};

export default Account;
