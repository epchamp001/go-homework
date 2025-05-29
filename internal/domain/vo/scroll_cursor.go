package vo

// ScrollCursor используется для реализации бесконечной прокрутки.
//
// LastID — идентификатор последнего элемента предыдущей порции,
// Limit — максимальное количество элементов в текущей порции.
type ScrollCursor struct {
	LastID string
	Limit  int
}
