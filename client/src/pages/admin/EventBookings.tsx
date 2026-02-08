import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { getEventBookings } from '../../api/bookings';
import { getEventById } from '../../api/events';
import type { BookingWithDetails, Event } from '../../types';

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    PENDING: 'bg-amber-100 text-amber-700',
    PAID: 'bg-green-100 text-green-700',
    REFUNDED: 'bg-purple-100 text-purple-700',
    CANCELLED: 'bg-red-100 text-red-700',
  };

  return (
    <span className={`px-2 py-1 text-xs font-medium rounded ${styles[status] || 'bg-neutral-100 text-neutral-700'}`}>
      {status}
    </span>
  );
}

export function EventBookings() {
  const { id } = useParams<{ id: string }>();
  const [event, setEvent] = useState<Event | null>(null);
  const [bookings, setBookings] = useState<BookingWithDetails[]>([]);
  const [loading, setLoading] = useState(true);
  const [status, setStatus] = useState('');
  const [sortBy, setSortBy] = useState('created_at');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');

  useEffect(() => {
    const fetchData = async () => {
      try {
        const eventData = await getEventById(Number(id));
        setEvent(eventData.event);
      } catch (error) {
        console.error('Failed to fetch event:', error);
      }
    };
    fetchData();
  }, [id]);

  useEffect(() => {
    const fetchBookings = async () => {
      setLoading(true);
      try {
        const data = await getEventBookings(Number(id), {
          status: status || undefined,
          sort: sortBy,
          order: sortOrder,
        });
        setBookings(data);
      } catch (error) {
        console.error('Failed to fetch bookings:', error);
      } finally {
        setLoading(false);
      }
    };
    fetchBookings();
  }, [id, status, sortBy, sortOrder]);

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
    <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <Link to="/admin/events" className="text-primary-500 hover:text-primary-600 text-sm mb-2 inline-block">
          ← Back to Events
        </Link>
        <h1 className="text-3xl font-bold text-neutral-900">
          Bookings for {event?.name || 'Event'}
        </h1>
      </div>

      <div className="bg-white rounded-xl border border-neutral-200 overflow-hidden">
        <div className="p-4 border-b border-neutral-200 flex flex-wrap gap-4 items-center">
          <div>
            <label className="block text-sm text-neutral-600 mb-1">Status</label>
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="px-3 py-2 border border-neutral-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="">All</option>
              <option value="PENDING">Pending</option>
              <option value="PAID">Paid</option>
              <option value="REFUNDED">Refunded</option>
              <option value="CANCELLED">Cancelled</option>
            </select>
          </div>

          <div>
            <label className="block text-sm text-neutral-600 mb-1">Sort By</label>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              className="px-3 py-2 border border-neutral-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="created_at">Date</option>
              <option value="status">Status</option>
            </select>
          </div>

          <div>
            <label className="block text-sm text-neutral-600 mb-1">Order</label>
            <select
              value={sortOrder}
              onChange={(e) => setSortOrder(e.target.value as 'asc' | 'desc')}
              className="px-3 py-2 border border-neutral-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="desc">Newest First</option>
              <option value="asc">Oldest First</option>
            </select>
          </div>

          <div className="ml-auto text-sm text-neutral-600">
            Total: {bookings.length} bookings
          </div>
        </div>

        {loading ? (
          <div className="flex justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
          </div>
        ) : bookings.length === 0 ? (
          <div className="text-center py-12 text-neutral-600">No bookings found</div>
        ) : (
          <table className="w-full">
            <thead className="bg-neutral-50">
              <tr>
                <th className="text-left py-3 px-4 text-sm font-medium text-neutral-600">ID</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-neutral-600">User</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-neutral-600">Status</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-neutral-600">Booked At</th>
              </tr>
            </thead>
            <tbody>
              {bookings.map((booking) => (
                <tr key={booking.booking_id} className="border-t border-neutral-100">
                  <td className="py-3 px-4 text-neutral-600">#{booking.booking_id}</td>
                  <td className="py-3 px-4">
                    <div className="font-medium text-neutral-900">{booking.user_name}</div>
                    <div className="text-sm text-neutral-500">{booking.user_email}</div>
                  </td>
                  <td className="py-3 px-4">
                    <StatusBadge status={booking.status} />
                  </td>
                  <td className="py-3 px-4 text-neutral-600 text-sm">
                    {formatDate(booking.created_at)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
