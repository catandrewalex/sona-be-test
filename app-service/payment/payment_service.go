package payment

import "time"

type EnrollmentPayment struct {
	ID           EnrollmentPaymentID `json:"id"`
	PaymentDate  time.Time           `json:"paymentDate"`
	BalanceTopUp int32               `json:"balanceTopUp"`
	Value        int32               `json:"value"`
	ValuePenalty int32               `json:"valuePenalty"`
}

type EnrollmentPaymentID int64

type PaymentService interface {
}
