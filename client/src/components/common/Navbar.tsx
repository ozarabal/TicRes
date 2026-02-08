import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { Button } from './Button';

export function Navbar() {
  const { user, isAuthenticated, isAdmin, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  return (
    <nav className="bg-white border-b border-neutral-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex items-center">
            <Link to="/" className="text-xl font-bold text-primary-500">
              TicRes
            </Link>
            <div className="hidden sm:ml-8 sm:flex sm:space-x-4">
              <Link
                to="/"
                className="text-neutral-600 hover:text-neutral-900 px-3 py-2 text-sm font-medium"
              >
                Events
              </Link>
              {isAdmin && (
                <>
                  <Link
                    to="/admin/events"
                    className="text-neutral-600 hover:text-neutral-900 px-3 py-2 text-sm font-medium"
                  >
                    Manage Events
                  </Link>
                  <Link
                    to="/admin/bookings"
                    className="text-neutral-600 hover:text-neutral-900 px-3 py-2 text-sm font-medium"
                  >
                    All Bookings
                  </Link>
                </>
              )}
            </div>
          </div>

          <div className="flex items-center space-x-4">
            {isAuthenticated ? (
              <>
                <Link
                  to="/profile"
                  className="text-neutral-600 hover:text-neutral-900 text-sm font-medium"
                >
                  {user?.name}
                </Link>
                <Button variant="ghost" size="sm" onClick={handleLogout}>
                  Logout
                </Button>
              </>
            ) : (
              <>
                <Link to="/login">
                  <Button variant="ghost" size="sm">
                    Login
                  </Button>
                </Link>
                <Link to="/register">
                  <Button size="sm">Register</Button>
                </Link>
              </>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}
