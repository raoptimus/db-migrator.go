package console

type Dummy struct {
	confirm bool
}

func NewDummy(confirm bool) *Dummy {
	return &Dummy{confirm: confirm}
}

func (c *Dummy) Confirm(s string) bool {
	return c.confirm
}

func (c *Dummy) Info(message string) {
	return
}
func (c *Dummy) InfoLn(message string) {
	return
}
func (c *Dummy) Infof(message string, a ...any) {
	return
}

func (c *Dummy) Success(message string) {
	return
}
func (c *Dummy) SuccessLn(message string) {
	return
}
func (c *Dummy) Successf(message string, a ...any) {
	return
}

func (c *Dummy) Warn(message string) {
	return
}
func (c *Dummy) WarnLn(message string) {
	return
}
func (c *Dummy) Warnf(message string, a ...any) {
	return
}

func (c *Dummy) Error(message string) {
	return
}
func (c *Dummy) ErrorLn(message string) {
	return
}
func (c *Dummy) Errorf(message string, a ...any) {
	return
}

func (c *Dummy) Fatal(err error) {
	return
}

func (c *Dummy) NumberPlural(count int, one, many string) string {
	if count > 1 {
		return many
	}

	return one
}
