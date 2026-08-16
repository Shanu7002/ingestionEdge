package domain

type IngestionPayload struct {
	Priority   int   // 0 = metrics / 1 = alerts
	IngestedAt int64 // Unix nanoseconds (Edge timestamp)
	Data       []byte
	Release    func()
}
