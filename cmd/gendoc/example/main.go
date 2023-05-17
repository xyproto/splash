package main

import "fmt"

// calculateSquares calculates the sum of the squares of the digits of the given number
// and sends the result to the squareop channel.
func calculateSquares(number int, squareop chan int) {
	sum := 0
	for number != 0 {
		digit := number % 10
		sum += digit * digit
		number /= 10
	}
	squareop <- sum
}

// calculateCubes calculates the sum of the cubes of the digits of the given number
// and sends the result to the cubeop channel.
func calculateCubes(number int, cubeop chan int) {
	sum := 0
	for number != 0 {
		digit := number % 10
		sum += digit * digit * digit
		number /= 10
	}
	cubeop <- sum
}

func main() {
	number := 589
	sqrch := make(chan int)
	cubech := make(chan int)

	// Start two goroutines to calculate the sum of squares and cubes of the digits.
	go calculateSquares(number, sqrch)
	go calculateCubes(number, cubech)

	// Receive the results from the channels and add them.
	squares, cubes := <-sqrch, <-cubech
	fmt.Println("Final result", squares+cubes)
}
