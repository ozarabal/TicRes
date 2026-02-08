import { useState, useEffect, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getEventById } from '../api/events';
import { bookSeats } from '../api/bookings';
import { useAuth } from '../context/AuthContext';
import { Button } from '../components/common/Button';
import type { EventWithSeats, Seat } from '../types';

const SEATS_PER_PAGE = 100;

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(amount);
}

function SeatSelector({
  seats,
  selectedSeats,
  onToggle,
}: {
  seats: Seat[];
  selectedSeats: number[];
  onToggle: (seatId: number) => void;
}) {
  return (
    <div className="grid grid-cols-10 gap-2">
      {seats.map((seat) => {
        const isSelected = selectedSeats.includes(seat.seat_id);
        const isBooked = seat.is_booked;

        return (
          <button
            key={seat.seat_id}
            onClick={() => !isBooked && onToggle(seat.seat_id)}
            disabled={isBooked}
            className={`w-8 h-8 rounded text-xs font-medium transition-colors ${
              isBooked
                ? 'bg-neutral-300 text-neutral-500 cursor-not-allowed'
                : isSelected
                ? 'bg-primary-500 text-white'
                : 'bg-neutral-100 text-neutral-700 hover:bg-primary-100'
            }`}
            title={`Seat ${seat.seat_number} - ${formatCurrency(seat.price)}`}
          >
            {seat.seat_number.split('-')[1]}
          </button>
        );
      })}
    </div>
  );
}

export function EventDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const [eventData, setEventData] = useState<EventWithSeats | null>(null);
  const [selectedSeats, setSelectedSeats] = useState<number[]>([]);
  const [loading, setLoading] = useState(true);
  const [booking, setBooking] = useState(false);
  const [error, setError] = useState('');
  const [seatPage, setSeatPage] = useState(1);

  useEffect(() => {
    const fetchEvent = async () => {
      try {
        const data = await getEventById(Number(id));
        setEventData(data);
      } catch {
        setError('Failed to load event');
      } finally {
        setLoading(false);
      }
    };
    fetchEvent();
  }, [id]);

  const totalSeatPages = eventData ? Math.ceil(eventData.seats.length / SEATS_PER_PAGE) : 1;

  const pagedSeats = useMemo(() => {
    if (!eventData) return [];
    const start = (seatPage - 1) * SEATS_PER_PAGE;
    return eventData.seats.slice(start, start + SEATS_PER_PAGE);
  }, [eventData, seatPage]);

  const selectedTotal = useMemo(() => {
    if (!eventData) return 0;
    return eventData.seats
      .filter((s) => selectedSeats.includes(s.seat_id))
      .reduce((sum, s) => sum + s.price, 0);
  }, [eventData, selectedSeats]);

  const ticketPrice = eventData?.seats.find((s) => !s.is_booked)?.price ?? 0;

  // Count how many selected seats are on the current page
  const selectedOnPage = pagedSeats.filter((s) => selectedSeats.includes(s.seat_id)).length;

  const handleSeatToggle = (seatId: number) => {
    setSelectedSeats((prev) =>
      prev.includes(seatId) ? prev.filter((id) => id !== seatId) : [...prev, seatId]
    );
  };

  const handleBook = async () => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }

    if (selectedSeats.length === 0) {
      setError('Please select at least one seat');
      return;
    }

    setBooking(true);
    setError('');

    try {
      const result = await bookSeats({ event_id: Number(id), seat_ids: selectedSeats });
      navigate(`/payments/${result.booking_id}`);
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: string } } };
      setError(error.response?.data?.error || 'Booking failed');
    } finally {
      setBooking(false);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[50vh]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
      </div>
    );
  }

  if (!eventData) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-8">
        <p className="text-center text-neutral-600">Event not found</p>
      </div>
    );
  }

  const { event, seats } = eventData;
  const availableSeats = seats.filter((s) => !s.is_booked).length;

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="bg-white rounded-xl border border-neutral-200 overflow-hidden">
        <div className="p-6 border-b border-neutral-200">
          <div className="flex justify-between items-start mb-4">
            <h1 className="text-2xl font-bold text-neutral-900">{event.name}</h1>
            {event.status === 'cancelled' && (
              <span className="px-3 py-1 text-sm font-medium bg-red-100 text-red-600 rounded-full">
                Cancelled
              </span>
            )}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
            <div className="flex items-center text-neutral-600">
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              {event.location}
            </div>
            <div className="flex items-center text-neutral-600">
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              {formatDate(event.date)}
            </div>
            <div className="flex items-center text-neutral-600">
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              {availableSeats} / {event.capacity} seats available
            </div>
          </div>

          {ticketPrice > 0 && (
            <div className="mt-4 pt-4 border-t border-neutral-100">
              <span className="text-sm text-neutral-600">Ticket Price: </span>
              <span className="text-lg font-semibold text-neutral-900">{formatCurrency(ticketPrice)}</span>
              <span className="text-sm text-neutral-500"> / seat</span>
            </div>
          )}
        </div>

        {event.status !== 'cancelled' && (
          <div className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-neutral-900">Select Seats</h2>
              {totalSeatPages > 1 && (
                <span className="text-sm text-neutral-500">
                  Showing seats {(seatPage - 1) * SEATS_PER_PAGE + 1}-{Math.min(seatPage * SEATS_PER_PAGE, seats.length)} of {seats.length}
                </span>
              )}
            </div>

            <div className="mb-4 flex items-center space-x-6 text-sm">
              <div className="flex items-center">
                <div className="w-4 h-4 bg-neutral-100 rounded mr-2"></div>
                <span className="text-neutral-600">Available</span>
              </div>
              <div className="flex items-center">
                <div className="w-4 h-4 bg-primary-500 rounded mr-2"></div>
                <span className="text-neutral-600">Selected</span>
              </div>
              <div className="flex items-center">
                <div className="w-4 h-4 bg-neutral-300 rounded mr-2"></div>
                <span className="text-neutral-600">Booked</span>
              </div>
            </div>

            <div className="overflow-x-auto mb-4">
              <SeatSelector
                seats={pagedSeats}
                selectedSeats={selectedSeats}
                onToggle={handleSeatToggle}
              />
            </div>

            {/* Seat pagination */}
            {totalSeatPages > 1 && (
              <div className="flex items-center justify-center space-x-2 mb-6">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setSeatPage((p) => Math.max(1, p - 1))}
                  disabled={seatPage === 1}
                >
                  Previous
                </Button>
                {Array.from({ length: totalSeatPages }, (_, i) => i + 1).map((page) => {
                  const hasSelectedOnPage = eventData.seats
                    .slice((page - 1) * SEATS_PER_PAGE, page * SEATS_PER_PAGE)
                    .some((s) => selectedSeats.includes(s.seat_id));

                  return (
                    <button
                      key={page}
                      onClick={() => setSeatPage(page)}
                      className={`w-8 h-8 rounded text-sm font-medium transition-colors ${
                        page === seatPage
                          ? 'bg-primary-500 text-white'
                          : 'bg-neutral-100 text-neutral-700 hover:bg-neutral-200'
                      } ${hasSelectedOnPage && page !== seatPage ? 'ring-2 ring-primary-300' : ''}`}
                    >
                      {page}
                    </button>
                  );
                })}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setSeatPage((p) => Math.min(totalSeatPages, p + 1))}
                  disabled={seatPage === totalSeatPages}
                >
                  Next
                </Button>
              </div>
            )}

            {selectedOnPage > 0 && totalSeatPages > 1 && (
              <p className="text-xs text-neutral-500 text-center mb-4">
                {selectedOnPage} selected on this page, {selectedSeats.length} total across all pages
              </p>
            )}

            {error && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
                {error}
              </div>
            )}

            <div className="flex items-center justify-between">
              <div>
                <p className="text-neutral-600">
                  {selectedSeats.length} seat{selectedSeats.length !== 1 ? 's' : ''} selected
                </p>
                {selectedTotal > 0 && (
                  <p className="text-lg font-semibold text-neutral-900">
                    Total: {formatCurrency(selectedTotal)}
                  </p>
                )}
              </div>
              <Button onClick={handleBook} loading={booking} disabled={selectedSeats.length === 0}>
                {isAuthenticated ? 'Book Now' : 'Login to Book'}
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
