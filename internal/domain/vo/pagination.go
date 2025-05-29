// Package vo содержит value-object типы, используемые в качестве
// вспомогательных структур для передачи параметров между слоями.
package vo

// Pagination описывает параметры постраничной навигации.
//
// Page — номер страницы,
// Limit — количество элементов на страницу.
type Pagination struct {
	Page  int
	Limit int
}
