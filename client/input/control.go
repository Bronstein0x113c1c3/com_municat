package input

func (i *Input) Close() {
	close(i.signal)
}
func (i *Input) Play() {
	i.stream.Start()
}
func (i *Input) Mute() {
	i.stream.Abort()
}
