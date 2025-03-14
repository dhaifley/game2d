import { createBrowserRouter, RouteObject } from 'react-router-dom';
import Home from './pages/Home';
import NotFound from './pages/NotFound';

// Define the routes
const routes: RouteObject[] = [
  {
    path: '/',
    element: <Home />,
  },
  {
    path: '*',
    element: <NotFound />,
  },
];

// Create the router
const router = createBrowserRouter(routes);

export default router;
