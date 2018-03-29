package reflector

func ConcurrentInternalListenAndServe() chan error {
	ch := make(chan error)
	go func() {
		// internal logic here
	}()
	return ch
}
