import React, { useEffect, useState } from 'react';
import { Account as AccountType, fetchAccount, updateAccount, setAccountRepo, setAccountAIKey } from '../services/accountService';
import { useAuth } from '../contexts/AuthContext';
import Modal from './Modal';

// Helper function to check if a user has the required scope
const hasScope = (scopes: string | undefined, requiredScope: string): boolean => {
  if (!scopes) return false;
  return scopes.split(' ').includes(requiredScope) || scopes.split(' ').includes('superuser');
};

interface AccountProps {
  onClose: () => void;
}

const Account: React.FC<AccountProps> = ({ onClose }) => {
  const { user: authUser } = useAuth();
  
  // Component state
  const [account, setAccount] = useState<AccountType | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Edit mode state
  const [isEditMode, setIsEditMode] = useState(false);
  const [editedName, setEditedName] = useState('');
  const [editedMaxTokens, setEditedMaxTokens] = useState<number | undefined>(undefined);
  const [editedThinkingBudget, setEditedThinkingBudget] = useState<string | undefined>(undefined);
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  
  // Repo modal state
  const [isRepoModalOpen, setIsRepoModalOpen] = useState(false);
  const [repoUrl, setRepoUrl] = useState('');
  const [isRepoSaving, setIsRepoSaving] = useState(false);
  const [repoError, setRepoError] = useState<string | null>(null);
  
  // AI Key modal state
  const [isAIKeyModalOpen, setIsAIKeyModalOpen] = useState(false);
  const [aiKey, setAIKey] = useState('');
  const [isAIKeySaving, setIsAIKeySaving] = useState(false);
  const [aiKeyError, setAIKeyError] = useState<string | null>(null);

  // Check if the current user has admin permission
  const canEdit = authUser && hasScope(authUser.scopes, 'account:admin');

  // Load account data when component mounts
  useEffect(() => {
    const loadAccount = async () => {
      try {
        setLoading(true);
        setError(null);
        const accountData = await fetchAccount();
        setAccount(accountData);
        setEditedName(accountData.name || '');
        setEditedMaxTokens(accountData.ai_max_tokens);
        setEditedThinkingBudget(accountData.ai_thinking_budget);
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
      setEditedMaxTokens(account.ai_max_tokens);
      setEditedThinkingBudget(account.ai_thinking_budget);
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
        name: editedName.trim(),
        ai_max_tokens: editedMaxTokens,
        ai_thinking_budget: editedThinkingBudget
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
  
  // Handle setting repository URL
  const handleSetRepo = async () => {
    if (!repoUrl.trim()) {
      setRepoError('Repository URL is required');
      return;
    }
    
    try {
      setIsRepoSaving(true);
      setRepoError(null);
      
      // Call the API to set the repository URL
      const updatedAccount = await setAccountRepo(repoUrl.trim());
      
      // Close the modal
      setIsRepoModalOpen(false);
      setRepoUrl('');
      
      // Update the local state with the updated account data
      setAccount(updatedAccount);
    } catch (err) {
      setRepoError(err instanceof Error ? err.message : 'Failed to set repository URL');
      console.error('Error setting repository URL:', err);
    } finally {
      setIsRepoSaving(false);
    }
  };
  
  // Handle setting AI API key
  const handleSetAIKey = async () => {
    if (!aiKey.trim()) {
      setAIKeyError('AI API Key is required');
      return;
    }
    
    try {
      setIsAIKeySaving(true);
      setAIKeyError(null);
      
      // Call the API to set the AI API key
      const updatedAccount = await setAccountAIKey(aiKey.trim());
      
      // Close the modal
      setIsAIKeyModalOpen(false);
      setAIKey('');
      
      // Update the local state with the updated account data
      setAccount(updatedAccount);
    } catch (err) {
      setAIKeyError(err instanceof Error ? err.message : 'Failed to set AI API key');
      console.error('Error setting AI API key:', err);
    } finally {
      setIsAIKeySaving(false);
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
            canEdit && (
              <>
                <button 
                  className="repo-button" 
                  onClick={() => {
                    setRepoUrl('');
                    setRepoError(null);
                    setIsRepoModalOpen(true);
                  }}
                >
                  Set Repo
                </button>
                <button 
                  className="ai-key-button" 
                  onClick={() => {
                    setAIKey('');
                    setAIKeyError(null);
                    setIsAIKeyModalOpen(true);
                  }}
                >
                  Set AI Key
                </button>
                <button className="edit-button" onClick={handleEditClick}>Edit</button>
              </>
            )
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
            <label htmlFor='game_commit_hash'>Game Commit:</label>
            <input
              type="text"
              id="game_commit_hash"
              value={account.game_commit_hash || ''}
              className="readonly-input"
              readOnly
            />
          </div>
          
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

          <div className="account-field-container">
            <label htmlFor='ai_max_tokens'>AI Max Tokens:</label>
            <input
              type="number"
              id="ai_max_tokens"
              value={isEditMode ? editedMaxTokens?.toString() || '' : account.ai_max_tokens?.toString() || ''}
              onChange={(e) => setEditedMaxTokens(e.target.value ? parseInt(e.target.value, 10) : undefined)}
              className={isEditMode ? "name-input" : "readonly-input"}
              readOnly={!isEditMode}
            />
          </div>
          
          <div className="account-field-container">
            <label htmlFor='ai_thinking_budget'>Thinking Budget:</label>
            <input
              type="number"
              id="ai_thinking_budget"
              value={isEditMode ? editedThinkingBudget || '' : account.ai_thinking_budget || ''}
              onChange={(e) => setEditedThinkingBudget(e.target.value)}
              className={isEditMode ? "name-input" : "readonly-input"}
              readOnly={!isEditMode}
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
      
      {/* Repository URL Modal */}
      <Modal
        isOpen={isRepoModalOpen}
        onClose={() => {
          setIsRepoModalOpen(false);
          setRepoUrl('');
          setRepoError(null);
        }}
        title="Set Import Repository"
        actions={
          <>
            <button 
              className="cancel-button" 
              onClick={() => {
                setIsRepoModalOpen(false);
                setRepoUrl('');
                setRepoError(null);
              }}
            >
              Cancel
            </button>
            <button 
              className="action-button" 
              onClick={handleSetRepo}
              disabled={isRepoSaving}
            >
              {isRepoSaving ? 'Setting...' : 'Set'}
            </button>
          </>
        }
      >
        <div className="modal-form-group">
          <label htmlFor="repo-url-input">Repository Connection URL</label>
          <input
            id="repo-url-input"
            type="text"
            value={repoUrl}
            onChange={(e) => setRepoUrl(e.target.value)}
            placeholder="Enter repository URL"
            autoFocus
          />
        </div>
        {repoError && <div className="modal-error">{repoError}</div>}
      </Modal>
      
      {/* AI API Key Modal */}
      <Modal
        isOpen={isAIKeyModalOpen}
        onClose={() => {
          setIsAIKeyModalOpen(false);
          setAIKey('');
          setAIKeyError(null);
        }}
        title="Set AI API Key"
        actions={
          <>
            <button 
              className="cancel-button" 
              onClick={() => {
                setIsAIKeyModalOpen(false);
                setAIKey('');
                setAIKeyError(null);
              }}
            >
              Cancel
            </button>
            <button 
              className="action-button" 
              onClick={handleSetAIKey}
              disabled={isAIKeySaving}
            >
              {isAIKeySaving ? 'Setting...' : 'Set'}
            </button>
          </>
        }
      >
        <div className="modal-form-group">
          <label htmlFor="ai-key-input">AI API Key</label>
          <input
            id="ai-key-input"
            type="text"
            value={aiKey}
            onChange={(e) => setAIKey(e.target.value)}
            placeholder="Enter AI API key"
            autoFocus
          />
        </div>
        {aiKeyError && <div className="modal-error">{aiKeyError}</div>}
      </Modal>
    </div>
  );
};

export default Account;
