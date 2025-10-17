package tasks

func (t *Tasks) GetWaitingCount() int {
	return t.waiting.GetCount()
}

func (t *Tasks) GetFinishedCount() int {
	return t.finished.GetCount()
}

func (t *Tasks) GetAddedTotalCount() int {
	return int(t.addedTotalCount.Load())
}

func (t *Tasks) GetReAddedTotalCount() int {
	return int(t.reAddedTotalCount.Load())
}

func (t *Tasks) GetTookTotalCount() int {
	return int(t.tookTotalCount.Load())
}
func (t *Tasks) GetFinishedTotalCount() int {
	return int(t.finishedTotalCount.Load())
}

func (t *Tasks) GetSuccessTotalCount() int {
	return int(t.successTotalCount.Load())
}

func (t *Tasks) GetErrorTotalCount() int {
	return int(t.errorTotalCount.Load())
}

func (t *Tasks) GetTimeoutTotalCount() int {
	return int(t.timeoutTotalCount.Load())
}
