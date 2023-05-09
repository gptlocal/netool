package io

type PipeReader struct {
	p *pipe
}

func (r *PipeReader) Read(data []byte) (n int, err error) {
	return r.p.read(data)
}

func (r *PipeReader) Close() error {
	return r.CloseWithError(nil)
}

func (r *PipeReader) CloseWithError(err error) error {
	return r.p.closeRead(err)
}
