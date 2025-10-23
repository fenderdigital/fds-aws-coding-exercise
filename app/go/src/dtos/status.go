package dtos

type SubStatus string

const (
	SubStatusActive    SubStatus = "active"
	SubStatusPending   SubStatus = "pending"
	SubStatusCancelled SubStatus = "cancelled"
)
