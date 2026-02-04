package mocks
import "github.com/stretchr/testify/mock"

type MockNotificationService struct {
	mock.Mock
}
func (m *MockNotificationService) SendNotification(bookingID int64, email, message string) {
	// Karena void function, kita cuma perlu rekam panggilan
	m.Called(bookingID, email, message)
}

func (m *MockNotificationService) EnqueueCancellation(eventID int64){
	m.Called(eventID)
}