import { createBrowserRouter, Navigate, RouteObject } from 'react-router-dom';
import Home from './pages/Home';
import NotFound from './pages/NotFound';
import Login from './pages/Login';
import Welcome from './pages/Welcome';
import Layout from './components/Layout';
import ProtectedRoute from './components/ProtectedRoute';

// Define the routes
const routes: RouteObject[] = [
  {
    path: '/',
    element: <Layout />,
    children: [
      {
        index: true,
        element: (
          <ProtectedRoute>
            <Home />
          </ProtectedRoute>
        ),
      },
      {
        path: 'login',
        element: <Login />
      },
      {
        path: 'welcome',
        element: <Welcome />
      },
      {
        path: '*',
        element: <NotFound />,
      },
    ],
  },
  // Redirect root access to welcome for unauthenticated users
  {
    path: '/',
    element: <Navigate to="/welcome" replace />
  }
];

// Create the router
const router = createBrowserRouter(routes);

export default router;
