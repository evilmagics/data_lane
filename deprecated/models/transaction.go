package models

import (
	"time"
)

type Transaction struct {
	Id          int       `db:"ID"`
	Branch      string    `db:"CB"`
	Gate        string    `db:"GB"`
	Station     string    `db:"GD"`
	Shift       string    `db:"SHIFT"`
	Period      string    `db:"PERIODA"`
	CollectorId string    `db:"IDPUL"`
	PasId       string    `db:"IDPAS"`
	Avc         string    `db:"AVC"`
	Datetime    time.Time `db:"WAKTU"`
	Class       string    `db:"GOL"`
	Method      string    `db:"METODA"`
	Serial      string    `db:"SERI"`
	Status      string    `db:"STATUS"`
	OriginGate  string    `db:"AG"`
	CardNumber  string    `db:"NOKARTU"`
	FirstImage  []byte    `db:"IMAGE1"`
	SecondImage []byte    `db:"IMAGE2"`
}

type TransactionFilter struct {
	StartDate   *time.Time
	EndDate     *time.Time
	TimeOverlap string
}

func (t Transaction) GetStation() string {
	if t.Station == "" {
		return "--"
	}
	return t.Station
}
func (t Transaction) GetShift() string {
	return t.Shift
}
func (t Transaction) GetPeriod() string {
	return t.Period
}
func (t Transaction) GetCollectorId() string {
	if t.CollectorId == "" {
		return "--"
	}
	return t.CollectorId
}
func (t Transaction) GetPasId() string {
	if t.PasId == "" {
		return "--"
	}
	return t.PasId
}
func (t Transaction) GetDatetime() string {
	return t.Datetime.Format("02-01-2006 15:04:05")
}
func (t Transaction) GetClass() string {
	if t.Class == "" {
		return "-"
	}
	return t.Class
}
func (t Transaction) GetAvc() string {
	return t.Avc
}
func (t Transaction) GetMethod() string {
	return TranslateTransactionMethod(t.Method)
}
func (t Transaction) GetSerial() string {
	if t.Serial == "" {
		return "000000"
	}
	return t.Serial
}
func (t Transaction) GetStatus() string {
	return t.Status
}
func (t Transaction) GetOriginGate() string {
	if t.OriginGate == "" {
		return "00"
	}
	return t.OriginGate
}
func (t Transaction) GetCardNumber() string {
	if t.CardNumber == "" {
		return "--"
	}
	return t.CardNumber
}

var (
	transactionMethodTranslation = map[string]string{
		"PPC":  "eToll Mdr",
		"PPC0": "eToll Mdr",
		"PPC1": "eToll BRI",
		"PPC2": "eToll BNI",
		"PPC3": "eToll BTN",
		"PPC5": "eToll BCA",
		"PPC7": "eToll DKI",
	}
)

func TranslateTransactionMethod(method string) string {
	if m, ok := transactionMethodTranslation[method]; ok {
		return m
	}
	return method
}
