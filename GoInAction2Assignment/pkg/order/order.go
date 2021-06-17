package order

import (
	pizza "GoInAction2Assignment/pkg/pizza"
	"fmt"
)

// GenerateOrderNo takes []OrderItem slice and checks if any OrderItem exists.
// If so, it will increment the orderNo by 1.
// A channel is use to receive the orderNo prevent multiple order no being generated at the same time
func GenerateOrderNo(orderNo int, orderSlice []OrderItem, ch chan<- int) {

	defer close(ch)

	// Increment newOrderNo global variable by 1 if there are OrderItem in the slice
	if len(orderSlice) > 0 {
		orderNo = orderNo + 1
	}

	ch <- orderNo
}

// CalOrderTotal calculates the total amount of an order
func CalOrderTotal(pizzaList *pizza.Linkedlist, orderSlice []OrderItem) float64 {

	orderTotal := 0.0
	pizzaTotal := 0.0

	var pizzaOrder pizza.Pizza

	for _, v := range orderSlice {
		pizzaOrder, _ = pizzaList.SearchPizza(v.PizzaNo)
		pizzaTotal = float64(v.OrderQty) * pizzaOrder.PizzaPrice
		orderTotal = orderTotal + pizzaTotal
	}

	return orderTotal
}

// PrintOrderReceipt prints the receipt of the order made on the server terminal/cmd prompt
func PrintOrderReceipt(pizzaList *pizza.Linkedlist, orderNo int, orderSlice []OrderItem, totalCost float64) {

	printDividerLine()
	fmt.Println("* RECEIPT *")
	fmt.Println()

	fmt.Println("Order No: ", orderNo)
	fmt.Println()

	pizzaTotal := 0.0

	for _, v := range orderSlice {
		pizzaOrder, _ := pizzaList.SearchPizza(v.PizzaNo)
		pizzaTotal = float64(v.OrderQty) * pizzaOrder.PizzaPrice

		fmt.Printf("%d x %s\t$%.2f\n", v.OrderQty, pizzaOrder.PizzaName, pizzaTotal)
	}

	fmt.Println("\t\t\t--------")
	fmt.Printf("TOTAL PAYMENT\t\t$%.2f\n", totalCost)
	fmt.Println("\t\t\t--------")
}

// printDividerLine prints a divider line to segregate sections for easy viewing
// when printing on the terminal/cmd prompt
func printDividerLine() {
	fmt.Println("------------------------------------------------------------")
}
