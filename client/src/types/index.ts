export interface User {
  user_id: number;
  name: string;
  username: string;
  email: string;
  role: 'admin' | 'user';
  created_at: string;
}

export interface Event {
  event_id: number;
  name: string;
  location: string;
  date: string;
  capacity: number;
  status?: 'available' | 'cancelled' | 'completed';
  created_at: string;
  updated_at: string;
}

export interface Seat {
  seat_id: number;
  event_id: number;
  seat_number: string;
  category: string;
  price: number;
  is_booked: boolean;
}

export interface Booking {
  booking_id: number;
  user_id: number;
  event_id: number;
  status: 'PENDING' | 'PAID' | 'REFUNDED' | 'CANCELLED' | 'EXPIRED';
  total_amount: number;
  expires_at?: string;
  created_at: string;
}

export interface Transaction {
  payment_id: number;
  amount: number;
  payment_method: string;
  booking_id: number;
  transaction_date: string;
  external_id: string;
  status: string;
}

export interface BookingWithPayment {
  booking_id: number;
  event_id: number;
  status: string;
  total_amount: number;
  expires_at?: string;
  transaction?: Transaction;
}

export interface PaymentRequest {
  booking_id: number;
  payment_method: 'credit_card' | 'bank_transfer' | 'e_wallet';
}

export interface BookingWithDetails {
  booking_id: number;
  user_id: number;
  user_name: string;
  user_email: string;
  event_id: number;
  event_name: string;
  status: 'PENDING' | 'PAID' | 'REFUNDED' | 'CANCELLED' | 'EXPIRED';
  created_at: string;
}

export interface EventWithSeats {
  event: Event;
  seats: Seat[];
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: {
    total: number;
    page: number;
    limit: number;
    hasMore: boolean;
  };
}
