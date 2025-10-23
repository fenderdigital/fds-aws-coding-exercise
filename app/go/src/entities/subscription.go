package entities

type SubscriptionItem struct {
	PK             string                 `dynamodbav:"pk"`
	SK             string                 `dynamodbav:"sk"`
	Type           string                 `dynamodbav:"type"`
	PlanSKU        string                 `dynamodbav:"planSku"`
	StartDate      string                 `dynamodbav:"startDate"`
	ExpiresAt      string                 `dynamodbav:"expiresAt"`
	CancelledAt    *string                `dynamodbav:"cancelledAt"`
	LastModifiedAt string                 `dynamodbav:"lastModifiedAt"`
	Attributes     map[string]interface{} `dynamodbav:"attributes"`
}

type Plan struct {
	PK             string   `dynamodbav:"pk"`
	SK             string   `dynamodbav:"sk"`
	Type           string   `dynamodbav:"type"`
	SKU            string   `dynamodbav:"-"`
	Name           string   `dynamodbav:"name"`
	Price          float64  `dynamodbav:"price"`
	Currency       string   `dynamodbav:"currency"`
	BillingCycle   string   `dynamodbav:"billingCycle"`
	Features       []string `dynamodbav:"features"`
	Status         string   `dynamodbav:"status"`
	LastModifiedAt string   `dynamodbav:"lastModifiedAt"`
}
