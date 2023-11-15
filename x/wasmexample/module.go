package main

//go:wasm-module CosmosSDK
//export add
func add(x, y int) int

// This function is exported to JavaScript, so can be called using
// exports.multiply() in JavaScript.
//
//export multiply
func multiply(x, y int) int {
	return add(x, y) * add(x, y)
}

//export __zeropb_alloc_page
func __zeropb_alloc_page() *byte {
	return nil
}

//export __zeropb_free_page
func __zeropb_free_page(*byte) {}

func main() {}
