package valueobject

import (
	"errors"
	"math"
)

type Money struct {
	amount   int64
	currency string
}

func NewMoney(amount float64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, errors.New("currency must not be empty")
	}
	cents := int64(math.Round(amount * 100))
	return Money{amount: cents, currency: currency}, nil
}

func NewMoneyFromCents(cents int64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, errors.New("currency must not be empty")
	}
	return Money{amount: cents, currency: currency}, nil
}

func (m Money) Amount() float64 {
	return float64(m.amount) / 100
}

func (m Money) Cents() int64 {
	return m.amount
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) IsZero() bool {
	return m.amount == 0
}

func (m Money) IsPositive() bool {
	return m.amount > 0
}
