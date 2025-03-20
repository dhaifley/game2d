import { createBrowserRouter, Navigate, RouteObject } from 'react-router-dom';
import NotFound from './pages/NotFound';
import Login from './pages/Login';
import Welcome from './pages/Welcome';
import Help from './pages/Help';
import Games from './pages/Games';
import User from './pages/User';
import Account from './pages/Account';
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
            <Games />
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
        path: 'help',
        element: <Help />
      },
      {
        path: 'user',
        element: (
          <ProtectedRoute>
            <User />
          </ProtectedRoute>
        ),
      },
      {
        path: 'account',
        element: (
          <ProtectedRoute>
            <Account />
          </ProtectedRoute>
        ),
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
