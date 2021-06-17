// Package pizza implements a LinkedList data structure to add/edit/delete pizzas.
package pizza

import (
	"errors"
	"fmt"
)

// Define a Pizza struct
type Pizza struct {
	PizzaNo    int
	PizzaName  string
	PizzaPrice float64
}

// Node item for the LinkedList is a Pizza struct
type Node struct {
	Item Pizza
	Next *Node
}

// LinkedList struct for the pizza menu
type Linkedlist struct {
	Head *Node
	Size int
}

// CreateStartMenu creates a standard pizza menu.
func (p *Linkedlist) CreateStartMenu(standardPizza []string, standardPrice float64) error {

	for pizzaNo, pizzaName := range standardPizza {
		pizzaNo = pizzaNo + 1
		p.AddPizza(pizzaNo, pizzaName, standardPrice)
	}

	return nil
}

// AddPizza creates a Pizza struct which is then added to the LinkedList node item.
func (p *Linkedlist) AddPizza(pizzaNo int, pizzaName string, pizzaPrice float64) error {

	newPizza := Pizza{
		PizzaNo:    pizzaNo,
		PizzaName:  pizzaName,
		PizzaPrice: pizzaPrice,
	}

	newNode := &Node{
		Item: newPizza,
		Next: nil,
	}

	if p.Head == nil {
		p.Head = newNode
	} else {
		currentNode := p.Head
		for currentNode.Next != nil {
			currentNode = currentNode.Next
		}
		currentNode.Next = newNode
	}
	p.Size++

	return nil
}

// EditPizza updates the Pizza struct in the LinkedList node item.
func (p *Linkedlist) EditPizza(pizzaNo int, pizzaName string, pizzaPrice float64) error {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(">>", err)
		}
	}()

	currentNode := p.Head

	if currentNode != nil {
		if pizzaNo == currentNode.Item.PizzaNo {
			currentNode.Item.PizzaName = pizzaName
			currentNode.Item.PizzaPrice = pizzaPrice
		} else {
			for currentNode.Next != nil {
				currentNode = currentNode.Next

				if pizzaNo == currentNode.Item.PizzaNo {
					currentNode.Item.PizzaName = pizzaName
					currentNode.Item.PizzaPrice = pizzaPrice
				}
			}
		}
	} else {
		panic("No pizza found")
	}
	return errors.New(">> Invalid pizza no")
}

// DeletePizza remove the node in the LinkedList where pizzaNo is found.
func (p *Linkedlist) DeletePizza(pizzaNo int) error {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(">>", err)
		}
	}()

	currentNode := p.Head

	if currentNode != nil {

		for i := 0; i < p.Size; i++ {
			if currentNode.Item.PizzaNo == pizzaNo {
				if i > 0 {
					prevNode := p.GetAt(i - 1)
					prevNode.Next = p.GetAt(i).Next
				} else {
					p.Head = currentNode.Next
				}
				p.Size--
				return nil
			}
			currentNode = currentNode.Next
		}
	} else {
		panic("No pizza found")
	}
	return errors.New(">> Invalid pizza no")
}

// GetAt finds the position where a node is located and returns the node pointer.
func (p *Linkedlist) GetAt(pos int) *Node {
	currentNode := p.Head
	if pos < 0 {
		return currentNode
	}
	if pos > (p.Size - 1) {
		return nil
	}
	for i := 0; i < pos; i++ {
		currentNode = currentNode.Next
	}
	return currentNode
}

// SearchPizza finds the Pizza struct in the LinkedList node item.
// It then returns the Pizza item.
func (p *Linkedlist) SearchPizza(pizzaNo int) (Pizza, error) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(">> panic:", err)
		}
	}()

	currentNode := p.Head

	if currentNode != nil {
		if pizzaNo == currentNode.Item.PizzaNo {
			return currentNode.Item, nil
		} else {
			for currentNode.Next != nil {
				currentNode = currentNode.Next

				if pizzaNo == currentNode.Item.PizzaNo {
					return currentNode.Item, nil
				}
			}
		}
	} else {
		panic("no pizza found")
	}
	return currentNode.Item, errors.New("no pizza found")
}

// GetAllPizza finds all the Pizza struct in the LinkedList node item and appends it to a pizzaSlice.
// It returns the pizzaSlice.
func (p *Linkedlist) GetAllPizza() ([]Pizza, error) {
	pizzaSlice := make([]Pizza, 0)

	currentNode := p.Head

	if currentNode == nil {
		return pizzaSlice, errors.New("no pizza on the menu today")
	}

	pizzaSlice = append(pizzaSlice, currentNode.Item)
	for currentNode.Next != nil {
		currentNode = currentNode.Next
		pizzaSlice = append(pizzaSlice, currentNode.Item)
	}

	return pizzaSlice, nil
}
