package judge

// Result 表示评测结果
type Result struct {
	// CPU时间使用量（毫秒）
	CpuTimeUsed int64
	// 实际运行时间（毫秒）
	RealTimeUsed int64
	// 内存使用量（字节）
	MemoryUsed int64
	// 退出信号
	Signal int
	// 运行时错误标记
	ReFlag bool
	// 是否发生运行时错误
	RuntimeError bool
	// 运行时错误信息
	RuntimeErrorMessage string
}

// NewResult 创建一个新的评测结果
func NewResult() *Result {
	return &Result{}
}

// SetRuntimeError 设置运行时错误
func (r *Result) SetRuntimeError(err error) {
	r.RuntimeError = true
	r.RuntimeErrorMessage = err.Error()
}

// IsSuccess 检查是否成功
func (r *Result) IsSuccess() bool {
	return !r.RuntimeError && r.Signal == 0
}

// GetStatus 获取评测状态
func (r *Result) GetStatus() string {
	if r.RuntimeError {
		return "Runtime Error"
	}
	if r.Signal != 0 {
		return "Signal Error"
	}
	return "Success"
}
