// Package order implements a Queue data structure to enqueue/dequeue orders in a FIFO order.
package order

import (
	"errors"
	"fmt"
	"sync"
)

// Define an OrderItem struct
type OrderItem struct {
	PizzaNo  int
	OrderQty int
}

// Define an Order struct. OrderSlice can contain more than 1 OrderItem
type Order struct {
	OrderNo    int
	OrderSlice []OrderItem
	TotalCost  float64
	UserName   string
}

// Node item for the Queue is an Order struct
type Node struct {
	Item Order
	Next *Node
}

// Queue struct for orders
type Queue struct {
	Front *Node
	Back  *Node
	Size  int
}

// Enqueue adds an order to the queue
func (p *Queue) Enqueue(orderNo int, orderSlice []OrderItem, totalCost float64, userName string) error {

	newOrder := Order{
		OrderNo:    orderNo,
		OrderSlice: orderSlice,
		TotalCost:  totalCost,
		UserName:   userName,
	}

	newNode := &Node{
		Item: newOrder,
		Next: nil,
	}
	if p.Front == nil {
		p.Front = newNode
	} else {
		p.Back.Next = newNode
	}
	p.Back = newNode
	p.Size++

	return nil
}

// Dequeue removes an order from the queue
func (p *Queue) Dequeue(orderChannel chan<- Order) {

	defer close(orderChannel)

	item := p.Front.Item

	if p.Size == 1 {
		p.Front = nil
		p.Back = nil
	} else {
		p.Front = p.Front.Next
	}
	p.Size--

	orderChannel <- item
}

// SearchOrder finds the Order struct in the LinkedList node item.
// It then returns the Order item.
func (p *Queue) SearchOrder(orderNo int) (Order, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(">> panic:", err)
		}
	}()

	currentNode := p.Front

	if currentNode != nil {
		if currentNode.Item.OrderNo == orderNo {
			return currentNode.Item, nil
		} else {
			for currentNode.Next != nil {
				currentNode = currentNode.Next
				if currentNode.Item.OrderNo == orderNo {
					return currentNode.Item, nil
				}
			}
		}
	} else {
		panic("no orders found")
	}

	return currentNode.Item, errors.New("no orders found")
}

// SearchPizzaInOrder finds the current node's OrderSlice and call the func SearchPizzaInSlice
// using pizzaNo given as parameters and returns true if found and false otherwise.
func (p *Queue) SearchPizzaInOrder(pizzaNo int) (bool, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(">> panic:", err)
		}
	}()

	currentNode := p.Front

	if currentNode != nil {
		orderSlice := currentNode.Item.OrderSlice

		// check if pizza is found in the slice
		if SearchPizzaInSlice(pizzaNo, orderSlice) {
			return true, nil
		} else {
			for currentNode.Next != nil {
				currentNode = currentNode.Next
				orderSlice = currentNode.Item.OrderSlice

				// check if pizza is found in the slice
				if SearchPizzaInSlice(pizzaNo, orderSlice) {
					return true, nil
				}
			}
		}
	} else {
		panic("no orders found")
	}
	return false, errors.New("no orders found")
}

// SearchPizzaInSlice takes in the pizzaNo and orderSlice as parameters and return true
// if pizzaNo is found in the slice. Otherwise, returns false.
func SearchPizzaInSlice(pizzaNo int, orderSlice []OrderItem) bool {
	if len(orderSlice) > 0 {
		for _, v := range orderSlice {
			if v.PizzaNo == pizzaNo {
				return true // if pizza is found in any orders, return true
			}
		}
	}
	return false
}

// UpdateOrder updates the current node item that matches the orderNo given with a new orderSlice and totalCost.
func (p *Queue) UpdateOrder(orderNo int, orderSlice []OrderItem, totalCost float64, wg *sync.WaitGroup, mutex *sync.Mutex) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(">> panic:", err)
		}
	}()

	defer wg.Done()

	// Mutex lock for updates
	mutex.Lock()

	currentNode := p.Front

	if currentNode != nil {
		if currentNode.Item.OrderNo == orderNo {
			currentNode.Item.OrderSlice = orderSlice
			currentNode.Item.TotalCost = totalCost
		} else {
			for currentNode.Next != nil {
				currentNode = currentNode.Next
				if currentNode.Item.OrderNo == orderNo {
					currentNode.Item.OrderSlice = orderSlice
					currentNode.Item.TotalCost = totalCost
				}
			}
		}
	} else {
		panic("no orders found")
	}

	mutex.Unlock()
}

// IsEmpty return true/false if the LinkedList is empty or not
func (p *Queue) IsEmpty() bool {
	return p.Size == 0
}

// GetAllOrders appends the current node Order item that belongs to a user into an Order slice.
// Admin user is allowe to get all the orders.
func (p *Queue) GetAllOrders(userName string, isAdmin bool) ([]Order, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(">> panic:", err)
		}
	}()

	orderList := make([]Order, 0)

	currentNode := p.Front

	if currentNode != nil {
		if !isAdmin {
			if currentNode.Item.UserName == userName {
				orderList = append(orderList, currentNode.Item)
			}
			for currentNode.Next != nil {
				currentNode = currentNode.Next
				if currentNode.Item.UserName == userName {
					orderList = append(orderList, currentNode.Item)
				}
			}

		} else {
			orderList = append(orderList, currentNode.Item)
			for currentNode.Next != nil {
				currentNode = currentNode.Next
				orderList = append(orderList, currentNode.Item)
			}
		}
	} else {
		panic("No orders found")
	}

	return orderList, nil
}
