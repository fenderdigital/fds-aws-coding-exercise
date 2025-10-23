package dtos

type SubStatus string

const (
	SubStatusActive    SubStatus = "ACTIVE"
	SubStatusInactive  SubStatus = "INACTIVE"
	SubStatusPending   SubStatus = "PENDING"
	SubStatusCancelled SubStatus = "CANCELLED"
)
