package main

import (
	"data-structures-and-algorithms/dynamicarrays"
	"fmt"
)

func main() {
	// Example usage
	arr := dynamicarrays.NewDynamicArray(2)

	fmt.Printf("Initial - Size: %d, Capacity: %d\n", arr.GetSize(), arr.GetCapacity())

	arr.PushBack(1)
	arr.PushBack(2)
	fmt.Printf("After 2 pushbacks - Size: %d, Capacity: %d\n", arr.GetSize(), arr.GetCapacity())

	arr.PushBack(3) // This will trigger resize
	fmt.Printf("After 3rd pushback - Size: %d, Capacity: %d\n", arr.GetSize(), arr.GetCapacity())

	fmt.Printf("Element at index 0: %d\n", arr.Get(0))
	fmt.Printf("Element at index 1: %d\n", arr.Get(1))
	fmt.Printf("Element at index 2: %d\n", arr.Get(2))

	arr.Set(1, 99)
	fmt.Printf("After setting index 1 to 99: %d\n", arr.Get(1))

	popped := arr.PopBack()
	fmt.Printf("Popped element: %d\n", popped)
	fmt.Printf("After popback - Size: %d, Capacity: %d\n", arr.GetSize(), arr.GetCapacity())
}
