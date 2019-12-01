package utils

// NewQueue returns a new queue with the given initial size.
func NewQueue(size int) *Queue {
	return &Queue{
		nodes: make([]*TxPublish, size),
		size:  size,
	}
}

// Queue is a basic FIFO queue based on a circular list that resizes as needed.
type Queue struct {
	nodes []*TxPublish
	size  int
	head  int
	tail  int
	count int
}

// Push adds a node to the queue.
func (q *Queue) Push(n *TxPublish) {
	if q.head == q.tail && q.count > 0 {
		nodes := make([]*TxPublish, len(q.nodes)+q.size)
		copy(nodes, q.nodes[q.head:])
		copy(nodes[len(q.nodes)-q.head:], q.nodes[:q.head])
		q.head = 0
		q.tail = len(q.nodes)
		q.nodes = nodes
	}
	q.nodes[q.tail] = n
	q.tail = (q.tail + 1) % len(q.nodes)
	q.count++
}

// Pop removes and returns a node from the queue in first to last order.
func (q *Queue) Pop() *TxPublish {
	if q.count == 0 {
		return nil
	}
	node := q.nodes[q.head]
	q.head = (q.head + 1) % len(q.nodes)
	q.count--
	return node
}

// Node of double linked list
type Node struct {
	Block BlockPublish
	Hash  ShaHash
	Prev  *Node
	Next  *Node
}

// List double linked list structure
type List struct {
	Size      int
	Tail      *Node
	Start     *Node
	Filenames []string
}

// InsertNode to double linked list
func (l *List) InsertNode(newNode *Node) {
	if l.Start == nil {
		l.Start = newNode
		l.Tail = newNode
	} else {
		l.Tail.Next = newNode
		newNode.Prev = l.Tail
		l.Tail = newNode
		l.Filenames = append(l.Filenames, newNode.Block.Transaction.Name)
	}
	l.Size++
}
