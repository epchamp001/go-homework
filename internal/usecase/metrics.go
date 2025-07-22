package usecase

type BussinesMetrics interface {
	IncOrdersAccepted()
	IncOrdersIssued()
	IncOrdersReturned()
}
