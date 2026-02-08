import api from './axios';
import type { PaymentRequest, Transaction, BookingWithPayment } from '../types';

export const processPayment = async (data: PaymentRequest): Promise<Transaction> => {
  const response = await api.post<{ message: string; data: Transaction }>('/payments', data);
  return response.data.data;
};

export const getPaymentStatus = async (bookingId: number): Promise<BookingWithPayment> => {
  const response = await api.get<{ data: BookingWithPayment }>(`/payments/${bookingId}`);
  return response.data.data;
};
