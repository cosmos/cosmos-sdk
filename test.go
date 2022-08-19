package main

import "fmt"

type sai struct {
	name string
	age  int
}

func main() {
	var My_map = make(map[float64][]string)
	fmt.Println(My_map)

	// As we already know that make() function
	// always returns a map which is initialized
	// So, we can add values in it
	My_map[1.3] = append(My_map[1.3], "Rohit")
	My_map[1.5] = append(My_map[1.3], "Sumit")
	fmt.Println(My_map)

	for k, val := range My_map {
		fmt.Println("Key ", k)
		fmt.Println("Val ", val)
	}

	var a sai
	a.age = 10
	a.name = "Sai"
	fmt.Println(a)
}
