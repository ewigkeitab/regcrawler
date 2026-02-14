package models

// Regulation represents a single regulation entry.
type Regulation struct {
	Title     string `json:"title"`
	Date      string `json:"date"`
	Category  string `json:"category"`
	Link      string `json:"link"`
	Content   string `json:"content"`
	Keypoints string `json:"keypoints,omitempty"`
}
