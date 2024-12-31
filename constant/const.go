package constant

// 代表运行结果
const (
	Fail = iota
	Success
)

// 代表运行结果评测中的错误
type ContentErr struct {
	Msg string
}

func (c *ContentErr) Error() string {
	return "content error " + c.Msg
}

type SystemErr struct {
	Msg string
}

func (s *SystemErr) Error() string {
	return "system error " + s.Msg
}
