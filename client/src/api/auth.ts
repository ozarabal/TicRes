import api from './axios';
import type { LoginRequest, RegisterRequest, User, BookingWithDetails } from '../types';

export const login = async (data: LoginRequest): Promise<string> => {
  const response = await api.post<{ token: string }>('/login', data);
  return response.data.token;
};

export const register = async (data: RegisterRequest): Promise<void> => {
  await api.post('/register', data);
};

export const getMe = async (): Promise<User> => {
  const response = await api.get<{ data: User }>('/me');
  return response.data.data;
};

export const getMyBookings = async (): Promise<BookingWithDetails[]> => {
  const response = await api.get<{ data: BookingWithDetails[] }>('/me/bookings');
  return response.data.data || [];
};
