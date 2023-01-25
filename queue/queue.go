package queue

type Item struct {
	Key   string
	Value uint32
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int      { return len(pq) }
func (pq PriorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Value < pq[j].Value
}

func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*Item))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}
