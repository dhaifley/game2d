import axios from 'axios';
import * as mockService from './mockGameService';

// Environment flag to use mock data (for development)
const USE_MOCK_DATA = false;

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
  ai_data?: any;
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
      const searchObj = { name: { $regex: searchQuery } };
      queryParams.append('search', JSON.stringify(searchObj));
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
    const response = await axios.get(`/api/v1/games/${id}`);
    return response.data;
  } catch (error) {
    console.error(`Error fetching game ${id}:`, error);
    throw error;
  }
};
