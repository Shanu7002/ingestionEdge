package domain

type IngestionPayload struct {
	Priority   int // 0 = metrics / 1 = alerts
	Sender     string
	IngestedAt int64
	Data       []byte
	Release    func()
}
