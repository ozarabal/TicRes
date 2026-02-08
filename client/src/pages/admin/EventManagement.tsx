import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getEvents, createEvent, updateEvent, deleteEvent } from '../../api/events';
import { Button } from '../../components/common/Button';
import { Input } from '../../components/common/Input';
import type { Event } from '../../types';

interface EventFormData {
  name: string;
  location: string;
  date: string;
  capacity: number;
  ticket_price: number;
}

function EventModal({
  isOpen,
  onClose,
  onSubmit,
  initialData,
  loading,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: EventFormData) => void;
  initialData?: Event;
  loading: boolean;
}) {
  const [formData, setFormData] = useState<EventFormData>({
    name: '',
    location: '',
    date: '',
    capacity: 100,
    ticket_price: 50000,
  });

  useEffect(() => {
    if (initialData) {
      const date = new Date(initialData.date);
      const formatted = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
      setFormData({
        name: initialData.name,
        location: initialData.location,
        date: formatted,
        capacity: initialData.capacity,
        ticket_price: 50000,
      });
    } else {
      setFormData({ name: '', location: '', date: '', capacity: 100, ticket_price: 50000 });
    }
  }, [initialData, isOpen]);

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(formData);
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-xl max-w-md w-full p-6">
        <h2 className="text-xl font-semibold text-neutral-900 mb-4">
          {initialData ? 'Edit Event' : 'Create Event'}
        </h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="Event Name"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
          />
          <Input
            label="Location"
            value={formData.location}
            onChange={(e) => setFormData({ ...formData, location: e.target.value })}
            required
          />
          <Input
            label="Date & Time (YYYY-MM-DD HH:MM)"
            value={formData.date}
            onChange={(e) => setFormData({ ...formData, date: e.target.value })}
            placeholder="2024-12-31 20:00"
            required
          />
          <Input
            label="Capacity"
            type="number"
            min={1}
            value={formData.capacity}
            onChange={(e) => setFormData({ ...formData, capacity: Number(e.target.value) })}
            required
          />
          <Input
            label="Ticket Price (IDR)"
            type="number"
            min={0}
            value={formData.ticket_price}
            onChange={(e) => setFormData({ ...formData, ticket_price: Number(e.target.value) })}
            required
          />
          <div className="flex justify-end space-x-3 pt-4">
            <Button type="button" variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" loading={loading}>
              {initialData ? 'Update' : 'Create'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

export function EventManagement() {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingEvent, setEditingEvent] = useState<Event | undefined>();
  const [submitting, setSubmitting] = useState(false);

  const fetchEvents = async () => {
    try {
      const response = await getEvents({ limit: 100 });
      setEvents(response.data || []);
    } catch (error) {
      console.error('Failed to fetch events:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchEvents();
  }, []);

  const handleCreate = () => {
    setEditingEvent(undefined);
    setModalOpen(true);
  };

  const handleEdit = (event: Event) => {
    setEditingEvent(event);
    setModalOpen(true);
  };

  const handleDelete = async (eventId: number) => {
    if (!confirm('Are you sure you want to cancel this event?')) return;
    try {
      await deleteEvent(eventId);
      fetchEvents();
    } catch (error) {
      console.error('Failed to delete event:', error);
    }
  };

  const handleSubmit = async (data: EventFormData) => {
    setSubmitting(true);
    try {
      if (editingEvent) {
        await updateEvent(editingEvent.event_id, data);
      } else {
        await createEvent(data);
      }
      setModalOpen(false);
      fetchEvents();
    } catch (error) {
      console.error('Failed to save event:', error);
    } finally {
      setSubmitting(false);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold text-neutral-900">Manage Events</h1>
        <Button onClick={handleCreate}>Create Event</Button>
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-neutral-200 overflow-hidden">
          <table className="w-full">
            <thead className="bg-neutral-50">
              <tr>
                <th className="text-left py-3 px-4 text-sm font-medium text-neutral-600">Event</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-neutral-600">Location</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-neutral-600">Date</th>
                <th className="text-left py-3 px-4 text-sm font-medium text-neutral-600">Capacity</th>
                <th className="text-right py-3 px-4 text-sm font-medium text-neutral-600">Actions</th>
              </tr>
            </thead>
            <tbody>
              {events.map((event) => (
                <tr key={event.event_id} className="border-t border-neutral-100">
                  <td className="py-3 px-4">
                    <div className="font-medium text-neutral-900">{event.name}</div>
                    {event.status === 'cancelled' && (
                      <span className="text-xs text-red-500">Cancelled</span>
                    )}
                  </td>
                  <td className="py-3 px-4 text-neutral-600">{event.location}</td>
                  <td className="py-3 px-4 text-neutral-600">{formatDate(event.date)}</td>
                  <td className="py-3 px-4 text-neutral-600">{event.capacity}</td>
                  <td className="py-3 px-4 text-right space-x-2">
                    <Link to={`/admin/events/${event.event_id}/bookings`}>
                      <Button variant="ghost" size="sm">
                        Bookings
                      </Button>
                    </Link>
                    <Button variant="ghost" size="sm" onClick={() => handleEdit(event)}>
                      Edit
                    </Button>
                    {event.status !== 'cancelled' && (
                      <Button
                        variant="danger"
                        size="sm"
                        onClick={() => handleDelete(event.event_id)}
                      >
                        Cancel
                      </Button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <EventModal
        isOpen={modalOpen}
        onClose={() => setModalOpen(false)}
        onSubmit={handleSubmit}
        initialData={editingEvent}
        loading={submitting}
      />
    </div>
  );
}
