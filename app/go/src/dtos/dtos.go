package dtos

type SubscriptionResponse struct {
	UserID         string                    `json:"userId"`
	SubscriptionID string                    `json:"subscriptionId"`
	Plan           *SubscriptionResponsePlan `json:"plan"`
	StartDate      string                    `json:"startDate"`
	ExpiresAt      string                    `json:"expiresAt"`
	CancelledAt    *string                   `json:"cancelledAt"`
	Status         SubStatus                 `json:"status"`
	Attributes     map[string]interface{}    `json:"attributes"`
}

type SubscriptionResponsePlan struct {
	SKU          string   `json:"sku"`
	Name         string   `json:"name"`
	Price        float64  `json:"price"`
	Currency     string   `json:"currency"`
	BillingCycle string   `json:"billingCycle"`
	Features     []string `json:"features"`
}
