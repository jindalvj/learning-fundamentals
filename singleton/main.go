package main

import (
	"fmt"
	"sync"
)

type singleton struct {
	data string
}

var instance *singleton
var once sync.Once

func GetInstance() *singleton {
	once.Do(func() {
		fmt.Println("Creating singleton")
		instance = &singleton{
			data: "test",
		}
	})
	return instance
}

func main() {
	s1 := GetInstance()
	s2 := GetInstance()

	fmt.Println(s1.data, s2.data)
	fmt.Println(s1 == s2)
}
