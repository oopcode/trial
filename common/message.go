package common

// AppMsg defines the message sent by producer and read by consumer.
type AppMsg struct {
	ID        int64  `json:"id"`
	Timestamp string `json:"timestamp"`
}

// Check performs sanity checks for message data.
func (m *AppMsg) Check() error {
	// TODO: implement.
	return nil
}
