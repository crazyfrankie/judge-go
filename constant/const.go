package constant

// 评测结果状态
const (
	// Success 表示评测成功
	Success = iota
	// Fail 表示评测失败
	Fail
)

// 评测错误类型
const (
	// ErrTypeContent 内容错误
	ErrTypeContent = "content"
	// ErrTypeSystem 系统错误
	ErrTypeSystem = "system"
	// ErrTypeRuntime 运行时错误
	ErrTypeRuntime = "runtime"
)

// ContentErr 内容错误
type ContentErr struct {
	Msg string
}

func (e *ContentErr) Error() string {
	return "content error: " + e.Msg
}

// SystemErr 系统错误
type SystemErr struct {
	Msg string
}

func (e *SystemErr) Error() string {
	return "system error: " + e.Msg
}

// RuntimeErr 运行时错误
type RuntimeErr struct {
	Msg string
}

func (e *RuntimeErr) Error() string {
	return "runtime error: " + e.Msg
}
