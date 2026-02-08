import { BrowserRouter, Routes, Route, Outlet } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { PrivateRoute, AdminRoute } from './routes/PrivateRoute';
import { Navbar } from './components/common/Navbar';
import { LoginPage } from './pages/LoginPage';
import { RegisterPage } from './pages/RegisterPage';
import { HomePage } from './pages/HomePage';
import { EventDetailPage } from './pages/EventDetailPage';
import { ProfilePage } from './pages/ProfilePage';
import { EventManagement } from './pages/admin/EventManagement';
import { BookingManagement } from './pages/admin/BookingManagement';
import { EventBookings } from './pages/admin/EventBookings';
import { PaymentPage } from './pages/PaymentPage';

function MainLayout() {
  return (
    <div className="min-h-screen bg-neutral-50">
      <Navbar />
      <Outlet />
    </div>
  );
}

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          {/* Auth pages (no navbar) */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />

          {/* Main layout with navbar */}
          <Route element={<MainLayout />}>
            {/* Public routes */}
            <Route path="/" element={<HomePage />} />
            <Route path="/events/:id" element={<EventDetailPage />} />

            {/* Protected routes (authenticated users) */}
            <Route element={<PrivateRoute />}>
              <Route path="/profile" element={<ProfilePage />} />
              <Route path="/payments/:bookingId" element={<PaymentPage />} />
            </Route>

            {/* Admin routes */}
            <Route element={<AdminRoute />}>
              <Route path="/admin/events" element={<EventManagement />} />
              <Route path="/admin/bookings" element={<BookingManagement />} />
              <Route path="/admin/events/:id/bookings" element={<EventBookings />} />
            </Route>
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;
