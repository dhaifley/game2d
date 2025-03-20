import React, { useEffect, useState } from 'react';
import { User as UserType, fetchUser, updateUser } from '../services/userService';

interface UserProps {
  onClose: () => void;
}

const User: React.FC<UserProps> = ({ onClose }) => {
  // Component state
  const [user, setUser] = useState<UserType | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Edit mode state
  const [isEditMode, setIsEditMode] = useState(false);
  const [editedLastName, setEditedLastName] = useState('');
  const [editedFirstName, setEditedFirstName] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  // Load user data when component mounts
  useEffect(() => {
    const loadUser = async () => {
      try {
        setLoading(true);
        setError(null);
        const userData = await fetchUser();
        setUser(userData);
        setEditedLastName(userData.last_name || '');
        setEditedFirstName(userData.first_name || '');
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load user data');
        console.error('Error loading user:', err);
      } finally {
        setLoading(false);
      }
    };

    loadUser();
  }, []);

  // Handle entering edit mode
  const handleEditClick = () => {
    setIsEditMode(true);
    setSaveError(null);
  };

  // Handle canceling edit mode (restore original values)
  const handleCancelEdit = () => {
    setIsEditMode(false);
    if (user) {
      setEditedLastName(user.last_name || '');
      setEditedFirstName(user.first_name || '');
    }
    setSaveError(null);
  };

  // Handle saving edited user
  const handleSaveEdit = async () => {
    if (!user) return;
    
    try {
      setIsSaving(true);
      setSaveError(null);
      
      // Prepare the update payload
      const updates: Partial<UserType> = {
        last_name: editedLastName.trim(),
        first_name: editedFirstName.trim()
      };
      
      // Call the API to update the user
      const updatedUser = await updateUser(updates);
      
      // Exit edit mode
      setIsEditMode(false);
      
      // Update the local state with the updated user data
      setUser(updatedUser);
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'Failed to save changes');
      console.error('Error updating user:', err);
    } finally {
      setIsSaving(false);
    }
  };

  if (loading) {
    return <div className="loading-indicator">Loading user data...</div>;
  }

  if (error) {
    return <div className="error-message">Error: {error}</div>;
  }

  if (!user) {
    return <div className="error-message">No user data available</div>;
  }

  return (
    <div className="user-details-container">
      <div className="user-details-header">
        <div className="user-buttons">
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

      <div className="user-details-content">
        <div className="user-details-title">
          <h2>{user.id}</h2>
        </div>

        <div className="user-additional-fields">
          <div className="user-field-container">
            <label htmlFor='id'>ID:</label>
            <input
              type="text"
              id="id"
              value={user.id}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="user-field-container">
            <label htmlFor='email'>Email:</label>
            <input
              type="text"
              id="email"
              value={user.email || ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="user-field-container">
            <label htmlFor='last_name'>Last Name:</label>
            <input
              type="text"
              id="last_name"
              value={isEditMode ? editedLastName : (user.last_name || '')}
              onChange={(e) => setEditedLastName(e.target.value)}
              className={isEditMode ? "name-input" : "readonly-input"}
              readOnly={!isEditMode}
            />
          </div>
          <div className="user-field-container">
            <label htmlFor='first_name'>First Name:</label>
            <input
              type="text"
              id="first_name"
              value={isEditMode ? editedFirstName : (user.first_name || '')}
              onChange={(e) => setEditedFirstName(e.target.value)}
              className={isEditMode ? "name-input" : "readonly-input"}
              readOnly={!isEditMode}
            />
          </div>
          <div className="user-field-container">
            <label htmlFor='status'>Status:</label>
            <input
              type="text"
              id="status"
              value={user.status || ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="user-field-container">
            <label htmlFor='scopes'>Scopes:</label>
            <input
              type="text"
              id="scopes"
              value={user.scopes || ''}
              className="readonly-input"
              readOnly
            />
          </div>

          {user.data && (
            <div className="user-data-container">
              <h3>User Data:</h3>
              <pre className="user-data-json">
                {JSON.stringify(user.data, null, 2)}
              </pre>
            </div>
          )}

          <div className="user-field-container">
            <label htmlFor='created_at'>Created At:</label>
            <input
              type="text"
              id="created_at"
              value={user.created_at ? new Date(user.created_at * 1000).toLocaleString() : ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="user-field-container">
            <label htmlFor='created_by'>Created By:</label>
            <input
              type="text"
              id="created_by"
              value={user.created_by || ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="user-field-container">
            <label htmlFor='updated_at'>Updated At:</label>
            <input
              type="text"
              id="updated_at"
              value={user.updated_at ? new Date(user.updated_at * 1000).toLocaleString() : ''}
              className="readonly-input"
              readOnly
            />
          </div>
          <div className="user-field-container">
            <label htmlFor='updated_by'>Updated By:</label>
            <input
              type="text"
              id="updated_by"
              value={user.updated_by || ''}
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

export default User;
