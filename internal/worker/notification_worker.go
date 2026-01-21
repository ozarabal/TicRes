package worker

import (
	"log"
	"time"
)

type NotificationPayload struct {
	BookingID int64
	UserEmail string
	Message	  string
}

type NotificationWorker struct {
	JobQueue chan NotificationPayload
}

func NewNotificationWorker() *NotificationWorker {
	return &NotificationWorker{
		JobQueue: make(chan NotificationPayload, 100),
	}
}

func (w *NotificationWorker) Start() {
	go func() {
		log.Println("Notification Worker started...")

		for job := range w.JobQueue {
			w.processJob(job)
		}
	}()
}

func (w *NotificationWorker) processJob(job NotificationPayload){
	log.Printf("Mengirim Email ke %s untuk Booking ID %d...", job.UserEmail, job.BookingID)

	time.Sleep(2 * time.Second)

	log.Printf("email terkirim ke %s!", job.UserEmail)
}

func (w* NotificationWorker) SendNotification(bookingID int64, email, message string) {
	w.JobQueue <- NotificationPayload{
		BookingID: bookingID,
		UserEmail: email,
		Message: message,
	}
}