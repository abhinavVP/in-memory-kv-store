package ttlheap

type ExpiryItem struct {
	Key       string
	ExpiresAt uint64
	index     int 
}

type ExpiryHeap []*ExpiryItem

func (h ExpiryHeap) Len() int {
	return len(h)
}

func (h ExpiryHeap) Less(i, j int) bool {
	return h[i].ExpiresAt < h[j].ExpiresAt
}

func (h ExpiryHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *ExpiryHeap) Push(x any) {
	item := x.(*ExpiryItem)
	item.index = len(*h)
	*h = append(*h, item)
}

func (h *ExpiryHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}