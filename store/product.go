package store

import "time"

// User holds user-related info
type Product struct {
	ID          string    `json:"id" db:"pharmspaceid"`
	Name        string    `json:"name" db:"name"`
	Export      bool      `json:"export" db:"export"`
	Manufacture string    `json:"manufacture" db:"manufacture"`
	Timestamp   time.Time `json:"time" bson:"time" db:"timestamp"`
}

// PrepareUntrusted pre-processes a comment received from untrusted source by clearing all
// autogen fields and reset everything users not supposed to provide
func (f *Product) PrepareUntrusted() {
	f.ID = ""   // don't allow user to define ID, force auto-gen
	f.Name = "" // don't allow user to define ID, force auto-gen
	f.Export = false
	f.Timestamp = time.Time{} // reset time, force auto-gen
}
