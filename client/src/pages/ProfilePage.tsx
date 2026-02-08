import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { getMyBookings } from '../api/auth';
import { Button } from '../components/common/Button';
import type { BookingWithDetails } from '../types';

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    PENDING: 'bg-amber-100 text-amber-700',
    PAID: 'bg-green-100 text-green-700',
    REFUNDED: 'bg-purple-100 text-purple-700',
    CANCELLED: 'bg-red-100 text-red-700',
    EXPIRED: 'bg-neutral-100 text-neutral-600',
  };

  return (
    <span className={`px-2 py-1 text-xs font-medium rounded ${styles[status] || 'bg-neutral-100 text-neutral-700'}`}>
      {status}
    </span>
  );
}

export function ProfilePage() {
  const { user } = useAuth();
  const [bookings, setBookings] = useState<BookingWithDetails[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchBookings = async () => {
      try {
        const data = await getMyBookings();
        setBookings(data);
      } catch (error) {
        console.error('Failed to fetch bookings:', error);
      } finally {
        setLoading(false);
      }
    };
    fetchBookings();
  }, []);

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <h1 className="text-3xl font-bold text-neutral-900 mb-8">My Profile</h1>

      <div className="bg-white rounded-xl border border-neutral-200 p-6 mb-8">
        <h2 className="text-lg font-semibold text-neutral-900 mb-4">Account Information</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <p className="text-sm text-neutral-500">Name</p>
            <p className="text-neutral-900">{user?.name}</p>
          </div>
          <div>
            <p className="text-sm text-neutral-500">Email</p>
            <p className="text-neutral-900">{user?.email}</p>
          </div>
          <div>
            <p className="text-sm text-neutral-500">Role</p>
            <p className="text-neutral-900 capitalize">{user?.role}</p>
          </div>
          <div>
            <p className="text-sm text-neutral-500">Member Since</p>
            <p className="text-neutral-900">{user?.created_at ? formatDate(user.created_at) : '-'}</p>
          </div>
        </div>
      </div>

      <div className="bg-white rounded-xl border border-neutral-200 p-6">
        <h2 className="text-lg font-semibold text-neutral-900 mb-4">My Bookings</h2>

        {loading ? (
          <div className="flex justify-center py-8">
            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary-500"></div>
          </div>
        ) : bookings.length === 0 ? (
          <p className="text-neutral-600 text-center py-8">No bookings yet</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-neutral-200">
                  <th className="text-left py-3 px-4 text-sm font-medium text-neutral-600">Event</th>
                  <th className="text-left py-3 px-4 text-sm font-medium text-neutral-600">Status</th>
                  <th className="text-left py-3 px-4 text-sm font-medium text-neutral-600">Booked At</th>
                  <th className="text-right py-3 px-4 text-sm font-medium text-neutral-600">Action</th>
                </tr>
              </thead>
              <tbody>
                {bookings.map((booking) => (
                  <tr key={booking.booking_id} className="border-b border-neutral-100 last:border-0">
                    <td className="py-3 px-4 text-neutral-900">{booking.event_name}</td>
                    <td className="py-3 px-4">
                      <StatusBadge status={booking.status} />
                    </td>
                    <td className="py-3 px-4 text-neutral-600 text-sm">
                      {formatDate(booking.created_at)}
                    </td>
                    <td className="py-3 px-4 text-right">
                      {booking.status === 'PENDING' ? (
                        <Link to={`/payments/${booking.booking_id}`}>
                          <Button size="sm">Pay Now</Button>
                        </Link>
                      ) : booking.status === 'PAID' ? (
                        <Link to={`/payments/${booking.booking_id}`}>
                          <Button variant="ghost" size="sm">View Receipt</Button>
                        </Link>
                      ) : null}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
