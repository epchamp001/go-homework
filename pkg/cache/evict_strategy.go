package cache

type EvictStrategy[K comparable] interface {
	// RecordAccess вызывается при успешном Get или Set-обновлении.
	RecordAccess(K)
	// RecordInsertion вызывается при добавлении нового ключа. Если shouldEvict true, кэш обязан удалить evictK.
	RecordInsertion(K) (evictK K, shouldEvict bool)
	// RecordDeletion сообщает стратегии, что ключ удалён извне.
	RecordDeletion(K)
}
