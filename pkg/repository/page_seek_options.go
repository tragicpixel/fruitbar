package repository

// PageSeekOptions holds the information required too perform a seek operation against a paginated repository.
type PageSeekOptions struct {
	// The maximum number of records to return.
	RecordLimit int `json:"limit"`
	// The ID to begin the seek operation from.
	StartId uint `json:"startid"`
	// The direction to move away from the starting Id.
	Direction string `json:"direction"`
}

const (
	SeekDirectionAfter  = "after"
	SeekDirectionBefore = "before"
	SeekDirectionNone   = "none"
)
