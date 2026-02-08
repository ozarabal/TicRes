import api from './axios';
import type { BookingWithDetails, BookingWithPayment, PaginatedResponse } from '../types';

export interface BookSeatsData {
  event_id: number;
  seat_ids: number[];
}

export const bookSeats = async (data: BookSeatsData): Promise<BookingWithPayment> => {
  const response = await api.post<{ message: string; data: BookingWithPayment }>('/bookings', data);
  return response.data.data;
};

export interface GetBookingsParams {
  status?: string;
  sort?: string;
  order?: 'asc' | 'desc';
  page?: number;
  limit?: number;
}

export const getAllBookings = async (params: GetBookingsParams = {}): Promise<PaginatedResponse<BookingWithDetails>> => {
  const response = await api.get<PaginatedResponse<BookingWithDetails>>('/admin/bookings', { params });
  return response.data;
};

export const getEventBookings = async (
  eventId: number,
  params: Omit<GetBookingsParams, 'page' | 'limit'> = {}
): Promise<BookingWithDetails[]> => {
  const response = await api.get<{ data: BookingWithDetails[] }>(`/admin/events/${eventId}/bookings`, { params });
  return response.data.data || [];
};
