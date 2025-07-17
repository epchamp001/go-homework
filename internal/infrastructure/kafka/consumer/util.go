package consumer

import "github.com/twmb/franz-go/pkg/kgo"

// firstFetchErr возвращает первую ненулевую ошибку из fetch-результата, чтобы быстро прервать цикл.
func firstFetchErr(errs []kgo.FetchError) error {
	for i := range errs {
		if errs[i].Err != nil {
			return errs[i].Err
		}
	}
	return nil
}
