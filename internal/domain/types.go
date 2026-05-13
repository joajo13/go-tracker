package domain

// AssetType is the catalogue of asset types the agent recognises.
type AssetType string

const (
	AssetTypeCEDEAR  AssetType = "cedear"
	AssetTypeUSStock AssetType = "us_stock"
	AssetTypeCrypto  AssetType = "crypto"
	AssetTypeBond    AssetType = "bond"
	AssetTypeFX      AssetType = "fx"
)

// IsValid reports whether the AssetType matches one of the supported values.
func (a AssetType) IsValid() bool {
	switch a {
	case AssetTypeCEDEAR, AssetTypeUSStock, AssetTypeCrypto, AssetTypeBond, AssetTypeFX:
		return true
	default:
		return false
	}
}

// Currency is the ISO-like code persisted alongside every price and operation.
type Currency string

const (
	CurrencyARS Currency = "ARS"
	CurrencyUSD Currency = "USD"
)

// IsValid reports whether the currency is one we explicitly support today.
func (c Currency) IsValid() bool {
	switch c {
	case CurrencyARS, CurrencyUSD:
		return true
	default:
		return false
	}
}

// OperationType discriminates buy and sell operations.
type OperationType string

const (
	OperationBuy  OperationType = "BUY"
	OperationSell OperationType = "SELL"
)

// IsValid reports whether the operation type is one of the supported values.
func (o OperationType) IsValid() bool {
	return o == OperationBuy || o == OperationSell
}

// AlertStatus tracks the lifecycle of a fired alert.
type AlertStatus string

const (
	AlertStatusNew      AlertStatus = "new"
	AlertStatusSeen     AlertStatus = "seen"
	AlertStatusArchived AlertStatus = "archived"
)

// IsValid reports whether the alert status is one of the supported values.
func (s AlertStatus) IsValid() bool {
	switch s {
	case AlertStatusNew, AlertStatusSeen, AlertStatusArchived:
		return true
	default:
		return false
	}
}
