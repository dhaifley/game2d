// Used for base64 encoding the default SVG icon
const defaultSvgIcon = `
<svg width="32" height="32" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
  <rect width="32" height="32" rx="4" fill="#646cff"/>
  <path d="M8 16h16M16 8v16" stroke="white" stroke-width="2" stroke-linecap="round"/>
</svg>
`;

// Convert the SVG to base64 for use as an icon
const svgToBase64 = (svg: string) => {
  return window.btoa(svg);
};

// Generate a random date within the last year
const getRandomDate = () => {
  const now = new Date();
  const oneYearAgo = new Date();
  oneYearAgo.setFullYear(now.getFullYear() - 1);
  
  const randomTimestamp = oneYearAgo.getTime() + Math.random() * (now.getTime() - oneYearAgo.getTime());
  return Math.floor(randomTimestamp / 1000); // Convert to Unix timestamp (seconds)
};

// Generate random version string
const getRandomVersion = () => {
  const major = Math.floor(Math.random() * 3);
  const minor = Math.floor(Math.random() * 10);
  const patch = Math.floor(Math.random() * 20);
  return `${major}.${minor}.${patch}`;
};

// Users who might have updated games
const users = [
  'admin@game2d.ai',
  'developer@game2d.ai',
  'test-user',
  'designer@game2d.ai'
];

// Possible game statuses
const statuses = ['active', 'inactive', 'new'];

// Generate a mock game with the given index
const generateMockGame = (index: number) => {
  const hasIcon = Math.random() > 0.3; // 70% of games have an icon
  
  return {
    id: `game-${index}`,
    name: `Game ${index}`,
    version: getRandomVersion(),
    description: `This is a description for Game ${index}`,
    icon: hasIcon ? svgToBase64(defaultSvgIcon) : '',
    status: statuses[Math.floor(Math.random() * statuses.length)],
    source: "app",
    tags: ["test:test"],
    prompts: {"current":{"prompt": "test", "response": "test"}, data: {}},
    created_at: getRandomDate(),
    created_by: users[Math.floor(Math.random() * users.length)],
    updated_at: getRandomDate(),
    updated_by: users[Math.floor(Math.random() * users.length)],
  };
};

// Generate a list of mock games
const generateMockGames = (count: number) => {
  return Array(count).fill(null).map((_, i) => generateMockGame(i + 1));
};

// Store our mock games
const mockGames = generateMockGames(50);

// Mock fetch games API
export const fetchGames = async (
  searchQuery: string = '',
  size: number = 10,
  skip: number = 0,
  sort: { key: string, direction: 'asc' | 'desc' | null } = { key: 'name', direction: 'asc' }
) => {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 500));

  // Filter games based on search query
  let filteredGames = [...mockGames];
  
  if (searchQuery) {
    const lowerCaseQuery = searchQuery.toLowerCase();
    filteredGames = filteredGames.filter(game => 
      game.name.toLowerCase().includes(lowerCaseQuery) ||
      game.description.toLowerCase().includes(lowerCaseQuery)
    );
  }

  // Sort games
  if (sort.direction) {
    filteredGames.sort((a, b) => {
      const aValue = a[sort.key as keyof typeof a];
      const bValue = b[sort.key as keyof typeof b];
      
      if (!aValue && !bValue) return 0;
      if (!aValue) return 1;
      if (!bValue) return -1;
      
      if (typeof aValue === 'string' && typeof bValue === 'string') {
        return sort.direction === 'asc' 
          ? aValue.localeCompare(bValue)
          : bValue.localeCompare(aValue);
      }
      
      return sort.direction === 'asc' 
        ? (aValue < bValue ? -1 : 1)
        : (bValue < aValue ? -1 : 1);
    });
  }

  // Apply pagination
  const paginatedGames = filteredGames.slice(skip, skip + size);
  
  // Return the result with total count
  return {
    games: paginatedGames,
    total: filteredGames.length
  };
};

// Mock fetch single game API
export const fetchGame = async (id: string) => {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 300));
  
  const game = mockGames.find(g => g.id === id);
  
  if (!game) {
    throw new Error(`Game with ID ${id} not found`);
  }
  
  return game;
};
