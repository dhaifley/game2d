import axios from 'axios';

// Base API URL
const API_BASE_URL = '/api/v1';

// Define the user interface based on OpenAPI schema
export interface User {
  account_id: string;
  id: string;
  email: string;
  last_name: string;
  first_name: string;
  status: string;
  scopes: string;
  data?: any;
  created_at?: number;
  created_by?: string;
  updated_at?: number;
  updated_by?: string;
}

// Fetch current user
export const fetchUser = async (): Promise<User> => {
  try {
    const response = await axios.get(`${API_BASE_URL}/user`);
    return response.data;
  } catch (error) {
    console.error('Error fetching user:', error);
    throw error;
  }
};

// Update current user
export const updateUser = async (updates: Partial<User>): Promise<User> => {
  try {
    // For the User component, we only want to update last_name and first_name
    const payload = {
      last_name: updates.last_name,
      first_name: updates.first_name
    };
    
    const response = await axios.put(`${API_BASE_URL}/user`, payload);
    return response.data;
  } catch (error) {
    console.error('Error updating user:', error);
    throw error;
  }
};
