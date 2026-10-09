package main

import (
	"errors"
	"fmt"
	"sort"
)

type FoodItem struct {
	ID        string
	Name      string
	Price     int // Using int for money
	Category  string
	StockLeft int
}

type Order struct {
	OrderID string
	Items   map[string]int
}

// setupInventory keeps main() clean, just like your first script
func setupInventory() map[string]*FoodItem {
	return map[string]*FoodItem{
		"ITEM-1": {ID: "ITEM-1", Name: "Pho", Price: 50000, Category: "Main Dish", StockLeft: 10},
		"ITEM-2": {ID: "ITEM-2", Name: "Vietnamese Coffee", Price: 25000, Category: "Drink", StockLeft: 5},
		"ITEM-3": {ID: "ITEM-3", Name: "Banh Mi", Price: 30000, Category: "Main Dish", StockLeft: 8},
		"ITEM-4": {ID: "ITEM-4", Name: "CheeseBurger", Price: 20000, Category: "Main Dish", StockLeft: 12},
		"ITEM-5": {ID: "ITEM-5", Name: "Fried Chicken", Price: 40000, Category: "Main Dish", StockLeft: 7},
		"ITEM-6": {ID: "ITEM-6", Name: "Spring Rolls", Price: 15000, Category: "Appetizer", StockLeft: 0}, // Out of stock
	}
}

// searchFood brings back your instant lookup feature
func searchFood(inventory map[string]*FoodItem, itemID string) {
	food, exists := inventory[itemID]
	if !exists {
		fmt.Printf("Search: '%s' not found in store.\n", itemID)
	} else {
		fmt.Printf("Search: Found '%s' (%s) - Price: %d VND, Stock Left: %d\n",
			food.Name, food.Category, food.Price, food.StockLeft)
	}
}

// isValidQuantity prevents negative or zero orders
func isValidQuantity(quantity int) bool {
	return quantity > 0
}

func addItemToOrder(order *Order, inventory map[string]*FoodItem, itemID string, qty int) error {
	// 1. Check for valid positive quantity (from your first script)
	if !isValidQuantity(qty) {
		return fmt.Errorf("cannot order %d. Quantity must be > 0", qty)
	}

	// 2. Check if item exists
	food, exists := inventory[itemID]
	if !exists {
		return errors.New("item does not exist")
	}

	// 3. Check stock (replaces your Available boolean)
	if food.StockLeft < qty {
		return fmt.Errorf("not enough stock for %s (requested: %d, left: %d)", food.Name, qty, food.StockLeft)
	}

	// 4. Process order
	food.StockLeft -= qty
	order.Items[itemID] += qty
	return nil
}

func listInventory(inventory map[string]*FoodItem) {
	if len(inventory) == 0 {
		fmt.Println("Inventory is empty.")
		return
	}

	keys := make([]string, 0, len(inventory))
	for k := range inventory {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Println("=== Available Menu ===")
	for _, itemID := range keys {
		item := inventory[itemID]

		// Only display items that are actually in stock (similar to your Available check)
		if item.StockLeft > 0 {
			fmt.Printf("[%s] %-17s | Category: %-9s | Price: %5d VND | Stock: %d\n",
				itemID, item.Name, item.Category, item.Price, item.StockLeft)
		}
	}
}

func printOrderSummary(order Order, inventory map[string]*FoodItem) {
	fmt.Printf("\n=== Order Summary (%s) ===\n", order.OrderID)

	if len(order.Items) == 0 {
		fmt.Println("Order is empty.")
		return
	}

	var grandTotal int = 0

	for itemID, qty := range order.Items {
		food := inventory[itemID]
		subtotal := food.Price * qty
		grandTotal += subtotal

		fmt.Printf("- %-17s | Qty: %d x %d VND = %d VND\n", food.Name, qty, food.Price, subtotal)
	}

	fmt.Println("------------------------------------------------")
	fmt.Printf("Total Money: %d VND\n\n", grandTotal)
}

func main() {
	// Initialize store
	inventory := setupInventory()

	fmt.Println("--- WELCOME TO THE CAMPUS FOOD SHOP ---")
	listInventory(inventory)

	fmt.Println("\n--- TESTING SEARCH ---")
	searchFood(inventory, "ITEM-1")
	searchFood(inventory, "ITEM-99") // Doesn't exist

	fmt.Println("\n--- TESTING ADD TO ORDER ---")
	myOrder := Order{
		OrderID: "ORD-1001",
		Items:   make(map[string]int),
	}

	// Test 1: Normal valid order
	if err := addItemToOrder(&myOrder, inventory, "ITEM-1", 2); err != nil {
		fmt.Println("Error:", err)
	}

	// Test 2: Reject invalid quantity (restored feature)
	if err := addItemToOrder(&myOrder, inventory, "ITEM-2", -3); err != nil {
		fmt.Println("Error:", err)
	}

	// Test 3: Reject unavailable/out-of-stock (restored feature)
	if err := addItemToOrder(&myOrder, inventory, "ITEM-6", 1); err != nil {
		fmt.Println("Error:", err)
	}

	printOrderSummary(myOrder, inventory)
}
