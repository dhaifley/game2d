import axios from 'axios';

// Base API URL
const API_BASE_URL = '/api/v1';

// Define the account interface based on OpenAPI schema
export interface Account {
  id: string;
  name: string;
  status: string;
  status_data?: any;
  repo?: string;
  repo_status?: string;
  repo_status_data?: any;
  game_commit_hash?: string;
  game_limit?: number;
  ai_api_key?: string;
  ai_max_tokens?: number;
  ai_thinking_budget?: string;
  data?: any;
  created_at?: number;
  updated_at?: number;
}

// Fetch current account
export const fetchAccount = async (): Promise<Account> => {
  try {
    const response = await axios.get(`${API_BASE_URL}/account`);
    return response.data;
  } catch (error) {
    console.error('Error fetching account:', error);
    throw error;
  }
};

// Update current account
export const updateAccount = async (updates: Partial<Account>): Promise<Account> => {
  try {
    // According to requirements, we'll only update name and other specific fields as needed
    const payload = updates;
    
    const response = await axios.post(`${API_BASE_URL}/account`, payload);
    return response.data;
  } catch (error) {
    console.error('Error updating account:', error);
    throw error;
  }
};

// Set account repository URL
export const setAccountRepo = async (repoUrl: string): Promise<Account> => {
  try {
    const payload = {
      repo: repoUrl
    };
    
    const response = await axios.post(`${API_BASE_URL}/account`, payload);
    return response.data;
  } catch (error) {
    console.error('Error setting account repository:', error);
    throw error;
  }
};

// Set account AI API key
export const setAccountAIKey = async (apiKey: string): Promise<Account> => {
  try {
    const payload = {
      ai_api_key: apiKey
    };
    
    const response = await axios.post(`${API_BASE_URL}/account`, payload);
    return response.data;
  } catch (error) {
    console.error('Error setting AI API key:', error);
    throw error;
  }
};
