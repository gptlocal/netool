package io

type PipeWriter struct {
	p *pipe
}

func (w *PipeWriter) Write(data []byte) (n int, err error) {
	return w.p.write(data)
}

func (w *PipeWriter) Close() error {
	return w.CloseWithError(nil)
}

func (w *PipeWriter) CloseWithError(err error) error {
	return w.p.closeWrite(err)
}
