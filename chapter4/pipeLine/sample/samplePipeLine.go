package main

import "fmt"

func multiply(values []int, multiplier int) []int {
	multipliedValues := make([]int, len(values))
	for i, v := range values {
		multipliedValues[i] = v * multiplier
	}
	return multipliedValues
}

func multiplyStream(value, multiplier int) int {
	return value * multiplier
}

func add(values []int, additive int) []int {
	addedValues := make([]int, len(values))
	for i, v := range values {
		addedValues[i] = v + additive
	}
	return addedValues
}

func addStream(value, additive int) int {
	return value + additive
}

func main() {
	ints := []int{1, 2, 3, 4}
	for _, v := range multiply(add(multiply(ints, 2), 1), 2) {
		fmt.Println(v)
	}

	for _, v := range ints {
		fmt.Println(multiplyStream(addStream(multiplyStream(v, 2), 1), 2))
	}
}
