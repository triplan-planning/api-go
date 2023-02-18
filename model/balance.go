package model

type Balance struct {
	PositiveAmount uint32 `json:"positiveAmount"`
	NegativeAmount uint32 `json:"negativeAmount"`
	TotalAmount    int32  `json:"totalAmount"`
}
