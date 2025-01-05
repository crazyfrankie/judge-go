package judge

type Result struct {
	CpuTimeUsed  int
	RealTimeUsed int
	MemoryUsed   int
	Signal       int
	ReFlag       bool
	ReSyscall    int
}
