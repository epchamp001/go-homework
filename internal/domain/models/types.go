package models

// PackageType представляет допустимый тип упаковки для заказа.
type PackageType string

const (
	// PackageNone — заказ без упаковки.
	PackageNone PackageType = ""

	// PackageBag — пакет.
	PackageBag PackageType = "bag"

	// PackageBox — коробка.
	PackageBox PackageType = "box"

	// PackageFilm — плёнка.
	PackageFilm PackageType = "film"

	// PackageBagFilm — пакет + плёнка (ограничение пакета + наценка двух типов).
	PackageBagFilm PackageType = "bag+film"

	// PackageBoxFilm — коробка + плёнка (ограничение коробки + наценка двух типов).
	PackageBoxFilm PackageType = "box+film"
)

// PriceKopecks — тип для представления цены в копейках.
type PriceKopecks int64

const (
	// SurchargeBag — наценка за пакет: 5₽.
	SurchargeBag PriceKopecks = 500

	// SurchargeBox — наценка за коробку: 20₽.
	SurchargeBox PriceKopecks = 2000

	// SurchargeFilm — наценка за плёнку: 1₽.
	SurchargeFilm PriceKopecks = 100
)
