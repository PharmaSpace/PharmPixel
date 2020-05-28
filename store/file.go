package store

import "time"

// User holds user-related info
type File struct {
	ID        string    `json:"id" bson:"_id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Timestamp time.Time `json:"time" bson:"time"  db:"timestamp"`
}

// PrepareUntrusted pre-processes a comment received from untrusted source by clearing all
// autogen fields and reset everything users not supposed to provide
func (f *File) PrepareUntrusted() {
	f.ID = ""                 // don't allow user to define ID, force auto-gen
	f.Name = ""               // don't allow user to define ID, force auto-gen
	f.Timestamp = time.Time{} // reset time, force auto-gen
}
