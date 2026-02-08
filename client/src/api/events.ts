import api from './axios';
import type { Event, EventWithSeats, PaginatedResponse } from '../types';

export interface GetEventsParams {
  search?: string;
  page?: number;
  limit?: number;
}

export const getEvents = async (params: GetEventsParams = {}): Promise<PaginatedResponse<Event>> => {
  const response = await api.get<PaginatedResponse<Event>>('/events', { params });
  return response.data;
};

export const getEventById = async (id: number): Promise<EventWithSeats> => {
  const response = await api.get<{ data: EventWithSeats }>(`/events/${id}`);
  return response.data.data;
};

export interface CreateEventData {
  name: string;
  location: string;
  date: string;
  capacity: number;
  ticket_price?: number;
}

export const createEvent = async (data: CreateEventData): Promise<Event> => {
  const response = await api.post<Event>('/events', data);
  return response.data;
};

export const updateEvent = async (id: number, data: CreateEventData): Promise<Event> => {
  const response = await api.put<{ data: Event }>(`/admin/events/${id}`, data);
  return response.data.data;
};

export const deleteEvent = async (id: number): Promise<void> => {
  await api.delete(`/admin/events/${id}`);
};
