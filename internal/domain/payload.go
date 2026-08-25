package domain

// TODO: add the sender to track the alert/metric (this should be possible null)
type IngestionPayload struct {
	Priority   int // 0 = metrics / 1 = alerts
	IngestedAt int64
	Data       []byte
	Release    func()
}
