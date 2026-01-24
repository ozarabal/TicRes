package worker

import (
	"log"
	"time"
	"sync"	
)

type NotificationPayload struct {
	BookingID int64
	UserEmail string
	Message	  string
}

type NotificationWorker struct {
	JobQueue chan NotificationPayload
	wg sync.WaitGroup
}

func NewNotificationWorker() *NotificationWorker {
	return &NotificationWorker{
		JobQueue: make(chan NotificationPayload, 100),
	}
}

func (w *NotificationWorker) Start() {

	w.wg.Add(1)

	go func() {
		log.Println("Notification Worker started...")

		for job := range w.JobQueue {
			w.processJob(job)
		}

		log.Println("Notification Worker stopped")
	}()
}

func (w *NotificationWorker) processJob(job NotificationPayload){
	log.Printf("Mengirim Email ke %s untuk Booking ID %d...", job.UserEmail, job.BookingID)

	time.Sleep(3 * time.Second)

	log.Printf("email terkirim ke %s!", job.UserEmail)
}

func (w* NotificationWorker) SendNotification(bookingID int64, email, message string) {
	w.JobQueue <- NotificationPayload{
		BookingID: bookingID,
		UserEmail: email,
		Message: message,
	}
}

func (w *NotificationWorker) Stop() {
	log.Println("Stopping worker... processing remainings jobs...")
	close(w.JobQueue)
	w.wg.Wait()
	log.Println("All jobs finished. Worker safe to exit.")
}