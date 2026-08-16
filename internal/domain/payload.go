package domain

type IngestionPayload struct {
	Priority int // 0 = metrics / 1 = alerts
	Data     []byte
	Release  func()
}
