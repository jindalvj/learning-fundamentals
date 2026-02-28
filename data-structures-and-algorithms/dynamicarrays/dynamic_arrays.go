package dynamicarrays

//import "fmt"

type DynamicArray struct {
	capacity int
	size     int
	arr      []int
}

func NewDynamicArray(capacity int) *DynamicArray {
	return &DynamicArray{
		capacity: capacity,
		size:     0,
		arr:      make([]int, capacity),
	}
}

func (d *DynamicArray) Get(i int) int {
	return d.arr[i]
}

func (d *DynamicArray) Set(i int, val int) {
	d.arr[i] = val
}

func (d *DynamicArray) PushBack(val int) {
	if d.size == d.capacity {
		d.Resize()
	}

	d.arr[d.size] = val
	d.size++
}

func (d *DynamicArray) PopBack() int {
	d.size--
	return d.arr[d.size]
}

func (d *DynamicArray) Resize() {
	d.capacity = 2 * d.capacity
	newArray := make([]int, d.capacity)

	for i := 0; i < d.size; i++ {
		newArray[i] = d.arr[i]
	}

	d.arr = newArray
}

func (d *DynamicArray) GetSize() int {
	return d.size
}

func (d *DynamicArray) GetCapacity() int {
	return d.capacity
}
