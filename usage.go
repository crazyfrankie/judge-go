package judge

import "golang.org/x/sys/unix"

type MemoryStatus struct {
	VMSize int // 虚拟内存大小（VM Size）
	VMRSS  int // 常驻内存集大小（VM RSS）
	VMData int // 数据段大小（VmData）
	VMStk  int // 堆栈段大小（VmStk）
	VMExe  int // 可执行文件大小（VmExe）
	VMLib  int // 共享库占用内存大小（VmLib）
}

func MemoryUsage(fd int) (MemoryStatus, error) {
	ms := MemoryStatus{}
	body := make([]byte, 4096)
	count, err := unix.Pread(fd, body, 0)
	if err != nil {
		return ms, err
	}

	// Parse file contents
	for i := 0; i < count; i++ {
		switch body[i] {
		case 'V':
			i++
			if body[i] == 'm' {
				i++
				switch body[i] {
				case 'R': // VMRSS
					i += 2
					ms.VMRSS = extractMemoryValue(body[i:])
				case 'D': // VMData
					i += 2
					ms.VMData = extractMemoryValue(body[i:])
				case 'S': // VMSize, VMStk
					i++
					if body[i] == 't' {
						i++
						ms.VMStk = extractMemoryValue(body[i:])
					} else {
						ms.VMSize = extractMemoryValue(body[i:])
					}
				case 'E': // VMExe
					i += 2
					ms.VMExe = extractMemoryValue(body[i:])
				case 'L': // VmLib
					i++
					ms.VMLib = extractMemoryValue(body[i:])
				}
			}
		}
		// Skip to next line
		for body[i] != '\n' {
			i++
		}
	}

	return ms, nil
}

// Extract the number in the Vm row
func extractMemoryValue(body []byte) int {
	ans := 0
	for i := 0; i < len(body) && isDigit(body[i]); i++ {
		ans = ans*10 + int(body[i]-'0')
	}
	return ans
}

// Determine whether a byte is a number
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
