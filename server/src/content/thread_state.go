package content

import "execution"

type ThreadState struct {
	id int
	tc execution.ThreadContext
}
