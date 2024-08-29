package api

func RetryApiCall[T any, R any](f func(T) (*R, error), input T, retryAttempts int) (*R, error) {
	var err error
	var response *R
	for i := 0; i < retryAttempts; i++ {

		response, err = f(input)
		if err == nil {
			return response, nil
		}
	}
	return nil, err
}
