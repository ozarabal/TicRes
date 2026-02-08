import { useState, useEffect, useRef } from 'react';
import { useParams, Link } from 'react-router-dom';
import { getPaymentStatus, processPayment } from '../api/payments';
import { Button } from '../components/common/Button';
import type { BookingWithPayment, Transaction } from '../types';

const PAYMENT_METHODS = [
  { value: 'credit_card' as const, label: 'Credit Card', icon: '💳' },
  { value: 'bank_transfer' as const, label: 'Bank Transfer', icon: '🏦' },
  { value: 'e_wallet' as const, label: 'E-Wallet', icon: '📱' },
];

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(amount);
}

function CountdownTimer({ expiresAt }: { expiresAt: string }) {
  const [timeLeft, setTimeLeft] = useState('');
  const [isExpired, setIsExpired] = useState(false);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    const update = () => {
      const now = Date.now();
      const expires = new Date(expiresAt).getTime();
      const diff = expires - now;

      if (diff <= 0) {
        setTimeLeft('00:00');
        setIsExpired(true);
        if (intervalRef.current) clearInterval(intervalRef.current);
        return;
      }

      const minutes = Math.floor(diff / 60000);
      const seconds = Math.floor((diff % 60000) / 1000);
      setTimeLeft(`${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`);
    };

    update();
    intervalRef.current = setInterval(update, 1000);
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [expiresAt]);

  return (
    <div className={`text-center p-3 rounded-lg ${isExpired ? 'bg-red-50 text-red-600' : 'bg-amber-50 text-amber-700'}`}>
      <p className="text-sm font-medium">
        {isExpired ? 'Payment time expired' : 'Complete payment within'}
      </p>
      <p className="text-2xl font-bold font-mono">{timeLeft}</p>
    </div>
  );
}

function PaymentReceipt({ transaction }: { transaction: Transaction }) {
  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const methodLabel = PAYMENT_METHODS.find(m => m.value === transaction.payment_method)?.label || transaction.payment_method;

  return (
    <div className="bg-green-50 border border-green-200 rounded-xl p-6">
      <div className="text-center mb-6">
        <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-3">
          <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <h2 className="text-xl font-bold text-green-800">Payment Successful</h2>
        <p className="text-green-600 text-sm mt-1">Your booking has been confirmed</p>
      </div>

      <div className="space-y-3 text-sm">
        <div className="flex justify-between">
          <span className="text-neutral-600">Transaction ID</span>
          <span className="font-mono text-neutral-900">{transaction.external_id}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-neutral-600">Amount</span>
          <span className="font-semibold text-neutral-900">{formatCurrency(transaction.amount)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-neutral-600">Payment Method</span>
          <span className="text-neutral-900">{methodLabel}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-neutral-600">Date</span>
          <span className="text-neutral-900">{formatDate(transaction.transaction_date)}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-neutral-600">Status</span>
          <span className="px-2 py-0.5 text-xs font-medium bg-green-100 text-green-700 rounded">
            {transaction.status}
          </span>
        </div>
      </div>
    </div>
  );
}

export function PaymentPage() {
  const { bookingId } = useParams<{ bookingId: string }>();
  const [booking, setBooking] = useState<BookingWithPayment | null>(null);
  const [loading, setLoading] = useState(true);
  const [paying, setPaying] = useState(false);
  const [selectedMethod, setSelectedMethod] = useState<'credit_card' | 'bank_transfer' | 'e_wallet'>('credit_card');
  const [error, setError] = useState('');
  const [paidTransaction, setPaidTransaction] = useState<Transaction | null>(null);

  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const data = await getPaymentStatus(Number(bookingId));
        setBooking(data);
        if (data.transaction && data.transaction.status === 'COMPLETED') {
          setPaidTransaction(data.transaction);
        }
      } catch (err: unknown) {
        const error = err as { response?: { status?: number; data?: { error?: string } } };
        if (error.response?.status === 403) {
          setError('You do not have access to this booking.');
        } else if (error.response?.status === 404) {
          setError('Booking not found.');
        } else {
          setError('Failed to load payment details.');
        }
      } finally {
        setLoading(false);
      }
    };
    fetchStatus();
  }, [bookingId]);

  const handlePay = async () => {
    setPaying(true);
    setError('');

    try {
      const txn = await processPayment({
        booking_id: Number(bookingId),
        payment_method: selectedMethod,
      });
      setPaidTransaction(txn);
      setBooking(prev => prev ? { ...prev, status: 'PAID' } : prev);
    } catch (err: unknown) {
      const error = err as { response?: { status?: number; data?: { error?: string } } };
      if (error.response?.status === 410) {
        setError('Booking has expired. Please create a new booking.');
        setBooking(prev => prev ? { ...prev, status: 'EXPIRED' } : prev);
      } else if (error.response?.status === 409) {
        setError('Payment has already been completed for this booking.');
      } else {
        setError(error.response?.data?.error || 'Payment failed. Please try again.');
      }
    } finally {
      setPaying(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[50vh]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
      </div>
    );
  }

  if (error && !booking) {
    return (
      <div className="max-w-lg mx-auto px-4 py-12 text-center">
        <div className="bg-red-50 border border-red-200 rounded-xl p-6">
          <p className="text-red-600">{error}</p>
          <Link to="/" className="text-primary-500 hover:text-primary-600 text-sm mt-4 inline-block">
            Back to Events
          </Link>
        </div>
      </div>
    );
  }

  if (!booking) return null;

  const isExpired = booking.status === 'EXPIRED' ||
    (booking.expires_at && new Date(booking.expires_at).getTime() < Date.now() && booking.status === 'PENDING');

  return (
    <div className="max-w-lg mx-auto px-4 sm:px-6 py-8">
      <h1 className="text-2xl font-bold text-neutral-900 mb-6">Payment</h1>

      <div className="bg-white rounded-xl border border-neutral-200 overflow-hidden">
        {/* Booking Summary */}
        <div className="p-6 border-b border-neutral-200">
          <div className="flex justify-between items-center mb-2">
            <span className="text-sm text-neutral-600">Booking ID</span>
            <span className="font-mono text-neutral-900">#{booking.booking_id}</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-sm text-neutral-600">Total Amount</span>
            <span className="text-xl font-bold text-neutral-900">{formatCurrency(booking.total_amount)}</span>
          </div>
        </div>

        <div className="p-6">
          {/* Already paid - show receipt */}
          {paidTransaction && (
            <>
              <PaymentReceipt transaction={paidTransaction} />
              <div className="mt-6 flex justify-center">
                <Link to="/profile">
                  <Button variant="secondary">View My Bookings</Button>
                </Link>
              </div>
            </>
          )}

          {/* Expired */}
          {isExpired && !paidTransaction && (
            <div className="text-center">
              <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-3">
                <svg className="w-8 h-8 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h2 className="text-lg font-semibold text-neutral-900 mb-1">Booking Expired</h2>
              <p className="text-neutral-600 text-sm mb-4">
                Your payment window has closed. Please create a new booking.
              </p>
              <Link to={`/events/${booking.event_id}`}>
                <Button>Book Again</Button>
              </Link>
            </div>
          )}

          {/* Refunded */}
          {booking.status === 'REFUNDED' && !paidTransaction && (
            <div className="text-center">
              <div className="w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-3">
                <svg className="w-8 h-8 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                </svg>
              </div>
              <h2 className="text-lg font-semibold text-neutral-900 mb-1">Booking Refunded</h2>
              <p className="text-neutral-600 text-sm mb-4">
                This booking has been refunded due to event cancellation.
              </p>
              <Link to="/profile">
                <Button variant="secondary">View My Bookings</Button>
              </Link>
            </div>
          )}

          {/* Cancelled */}
          {booking.status === 'CANCELLED' && !paidTransaction && (
            <div className="text-center">
              <h2 className="text-lg font-semibold text-neutral-900 mb-1">Booking Cancelled</h2>
              <p className="text-neutral-600 text-sm mb-4">This booking has been cancelled.</p>
              <Link to="/">
                <Button variant="secondary">Browse Events</Button>
              </Link>
            </div>
          )}

          {/* Pending - show payment form */}
          {booking.status === 'PENDING' && !isExpired && !paidTransaction && (
            <>
              {booking.expires_at && (
                <div className="mb-6">
                  <CountdownTimer expiresAt={booking.expires_at} />
                </div>
              )}

              <div className="mb-6">
                <h3 className="text-sm font-medium text-neutral-700 mb-3">Select Payment Method</h3>
                <div className="space-y-2">
                  {PAYMENT_METHODS.map((method) => (
                    <label
                      key={method.value}
                      className={`flex items-center p-3 rounded-lg border-2 cursor-pointer transition-colors ${
                        selectedMethod === method.value
                          ? 'border-primary-500 bg-primary-50'
                          : 'border-neutral-200 hover:border-neutral-300'
                      }`}
                    >
                      <input
                        type="radio"
                        name="payment_method"
                        value={method.value}
                        checked={selectedMethod === method.value}
                        onChange={() => setSelectedMethod(method.value)}
                        className="sr-only"
                      />
                      <span className="text-xl mr-3">{method.icon}</span>
                      <span className="font-medium text-neutral-900">{method.label}</span>
                      {selectedMethod === method.value && (
                        <svg className="w-5 h-5 text-primary-500 ml-auto" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                        </svg>
                      )}
                    </label>
                  ))}
                </div>
              </div>

              {error && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
                  {error}
                </div>
              )}

              <Button onClick={handlePay} loading={paying} className="w-full" size="lg">
                Pay {formatCurrency(booking.total_amount)}
              </Button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
