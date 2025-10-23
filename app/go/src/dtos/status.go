package dtos

type SubStatus string

const (
	SubStatusActive    SubStatus = "active"
	SubStatusInactive  SubStatus = "inactive"
	SubStatusPending   SubStatus = "pending"
	SubStatusCancelled SubStatus = "cancelled"
)
