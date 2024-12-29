package judge

const (
	Fail = iota
	Success
)

func StdCheck(userInput, stdInput string) int {
	if len(userInput) != len(stdInput) {
		return Fail
	}

	for i := 0; i < len(userInput); i++ {
		if userInput[i] != stdInput[i] {
			return Fail
		}
	}

	return Success
}
