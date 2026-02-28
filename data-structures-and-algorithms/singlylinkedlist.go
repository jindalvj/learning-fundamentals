package main

type SinglyListNode struct {
	val  int
	next *SinglyListNode
}

type SinglyList struct {
	head *SinglyListNode
	len  int
}

func NewSinglyList() *SinglyList {
	return &SinglyList{}
}

func (list *SinglyList) Get(index int) *SinglyListNode {
	if index < 0 || index >= list.len {
		return nil
	}

	cur := list.head

	for i := 0; i < index; i++ {
		cur = cur.next
	}

	return cur
}

func (list *SinglyList) InsertHead(val int) {
	node := &SinglyListNode{val: val, next: list.head}
	list.head = node
	list.len++
}

func (list *SinglyList) InsertTail(val int) {
	node := &SinglyListNode{val: val}
	if list.head == nil {
		list.head = node
		list.len++
		return
	}

	cur := list.head
	for cur.next != nil {
		cur = cur.next
	}
	cur.next = node
	list.len++
}

func (list *SinglyList) Remove(index int) bool {
	if index < 0 || index >= list.len {
		return false
	}

	if index == 0 {
		list.head = list.head.next
		list.len--
		return true
	}

}
