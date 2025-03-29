import axios from 'axios';
import * as mockService from './mockGameService';

// Environment flag to use mock data (for development)
const USE_MOCK_DATA = false;

// Base API URL
const API_BASE_URL = '/api/v1';

// Define the game response interface
export interface Game {
  id: string;
  previous_id?: string;
  name: string;
  version?: string;
  description?: string;
  icon?: string;
  status?: string;
  status_data?: any;
  source?: string;
  tags?: string[];
  prompts?: any;
  created_at?: number;
  created_by?: string;
  updated_at?: number;
  updated_by?: string;
  [key: string]: any | undefined;
}

// Fetch games with optional filtering and pagination
export const fetchGames = async (
  searchQuery: string = '',
  size: number = 10,
  skip: number = 0,
  sort: { key: string, direction: 'asc' | 'desc' | null } = { key: 'name', direction: 'asc' }
): Promise<{ games: Game[], total: number }> => {
  if (USE_MOCK_DATA) {
    return mockService.fetchGames(searchQuery, size, skip, sort);
  }

  try {
    // Construct query params
    const queryParams = new URLSearchParams();
    queryParams.append('size', size.toString());
    queryParams.append('skip', skip.toString());

    // Add search param if there's a search query
    if (searchQuery) {
      queryParams.append('search', JSON.stringify(searchQuery));
    }

    // Add sorting if configured
    if (sort.direction) {
      const sortObj: Record<string, number> = {};
      sortObj[sort.key] = sort.direction === 'asc' ? 1 : -1;
      queryParams.append('sort', JSON.stringify(sortObj));
    }

    const response = await axios.get(`/api/v1/games?${queryParams.toString()}`);
    
    // Get total count from headers if available
    const totalCount = response.headers['x-total-count'] 
      ? parseInt(response.headers['x-total-count'], 10) 
      : response.data.length;
    
    return {
      games: response.data,
      total: totalCount
    };
  } catch (error) {
    console.error('Error fetching games:', error);
    throw error;
  }
};

// Fetch a single game by ID
export const fetchGame = async (id: string): Promise<Game> => {
  if (USE_MOCK_DATA) {
    return mockService.fetchGame(id);
  }

  try {
    const response = await axios.get(`${API_BASE_URL}/games/${id}?minimal=true`);
    return response.data;
  } catch (error) {
    console.error(`Error fetching game ${id}:`, error);
    throw error;
  }
};

// Create a new game
export const createGame = async (name: string): Promise<Game> => {
  if (USE_MOCK_DATA) {
    return mockService.fetchGame('mock-new-game');
  }

  try {
    const response = await axios.post(`${API_BASE_URL}/games`, { "name": name });
    return response.data;
  } catch (error) {
    console.error('Error creating game:', error);
    throw error;
  }
};

// Copy an existing game
export const copyGame = async (id: string, name: string): Promise<Game> => {
  if (USE_MOCK_DATA) {
    return mockService.fetchGame('mock-copy-game');
  }

  try {
    const response = await axios.post(`${API_BASE_URL}/games/copy`,
      { "id": id, "name": name });
    return response.data;
  } catch (error) {
    console.error(`Error copying game ${id}:`, error);
    throw error;
  }
};

// Delete a game
export const deleteGame = async (id: string): Promise<void> => {
  if (USE_MOCK_DATA) {
    return Promise.resolve();
  }

  try {
    await axios.delete(`${API_BASE_URL}/games/${id}`);
  } catch (error) {
    console.error(`Error deleting game ${id}:`, error);
    throw error;
  }
};

// Update a game
export const updateGame = async (id: string, updates: Partial<Game>): Promise<Game> => {
  if (USE_MOCK_DATA) {
    return mockService.fetchGame(id);
  }

  try {
    const response = await axios.put(`${API_BASE_URL}/games/${id}`, updates);
    return response.data;
  } catch (error) {
    console.error(`Error updating game ${id}:`, error);
    throw error;
  }
};
