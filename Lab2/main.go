package main

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type CustomerType string

const (
	Regular CustomerType = "Regular"
	Member  CustomerType = "Member"
)

type OrderType string

const (
	Pickup   OrderType = "Pickup"
	Delivery OrderType = "Delivery"
)

type FoodItem struct {
	ID        string
	Name      string
	Price     int // Using int for money
	Category  string
	StockLeft int
}

type Order struct {
	OrderID      string
	Items        map[string]int
	CustomerType CustomerType
	OrderType    OrderType
}

// NEW: Struct to hold the result of the preparation
type OrderResult struct {
	OrderID  string
	Status   string
	PrepTime time.Duration
}

const (
	memberDiscountPercent = 5
	deliveryFee           = 30000
)

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

// //=============================CALCULATION FUNCTIONS=============================////
func calculateSubtotal(order Order, inventory map[string]*FoodItem) int {
	subtotal := 0
	for itemID, qty := range order.Items {
		subtotal += inventory[itemID].Price * qty
	}
	return subtotal
}

// Discount applies to the food subtotal only, never the delivery fee.
func calculateDiscount(customerType CustomerType, foodSubtotal int) int {
	if customerType == Member {
		return foodSubtotal * memberDiscountPercent / 100
	}
	return 0
}

func calculateDeliveryFee(orderType OrderType) int {
	if orderType == Delivery {
		return deliveryFee
	}
	return 0
}

// ============================================================================//
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

// func printOrderSummary(order Order, inventory map[string]*FoodItem) {
// 	fmt.Printf("\n=== Order Summary (%s) ===\n", order.OrderID)

// 	if len(order.Items) == 0 {
// 		fmt.Println("Order is empty.")
// 		return
// 	}

// 	var grandTotal int = 0

// 	for itemID, qty := range order.Items {
// 		food := inventory[itemID]
// 		subtotal := food.Price * qty
// 		grandTotal += subtotal

// 		fmt.Printf("- %-17s | Qty: %d x %d VND = %d VND\n", food.Name, qty, food.Price, subtotal)
// 	}

//		fmt.Println("------------------------------------------------")
//		fmt.Printf("Total Money: %d VND\n\n", grandTotal)
//	}
func printOrderSummary(order Order, inventory map[string]*FoodItem) {
	fmt.Printf("\n=== Order Summary (%s) ===\n", order.OrderID)
	fmt.Printf("Customer: %s | Order type: %s\n", order.CustomerType, order.OrderType)

	if len(order.Items) == 0 {
		fmt.Println("Order is empty.")
		return
	}

	for itemID, qty := range order.Items {
		food := inventory[itemID]
		fmt.Printf("- %-17s | Qty: %d x %d VND = %d VND\n", food.Name, qty, food.Price, food.Price*qty)
	}

	subtotal := calculateSubtotal(order, inventory)
	discount := calculateDiscount(order.CustomerType, subtotal)
	fee := calculateDeliveryFee(order.OrderType)
	finalTotal := subtotal - discount + fee

	fmt.Println("------------------------------------------------")
	fmt.Printf("Subtotal:     %d VND\n", subtotal)
	fmt.Printf("Discount:    -%d VND\n", discount)
	fmt.Printf("Delivery fee: %d VND\n", fee)
	fmt.Printf("Final Total:  %d VND\n\n", finalTotal)
}

// ================Lab2 part====================//
// NEW: Function to simulate order preparation concurrently
// UPDATED: Worker function that returns a struct via channel
// 5.2 Step B — Concurrent Order Preparation with Channels
// func prepareOrder(orderID string, prepTime time.Duration, resultChan chan<- OrderResult) {
// 	startTime := time.Now()

// 	// Simulate preparation
// 	time.Sleep(prepTime)

// 	// Calculate exact time and send the result
// 	elapsed := time.Since(startTime)
// 	resultChan <- OrderResult{
// 		OrderID:  orderID,
// 		Status:   "READY",
// 		PrepTime: elapsed,
// 	}
// }

// Sequential version: No channels, no goroutines
func prepareOrderSequential(orderID string, prepTime time.Duration) {
	fmt.Printf("Order %s: Preparation started...\n", orderID)
	time.Sleep(prepTime)
	fmt.Printf("Order %s: Preparation completed!\n", orderID)
}

// 5.3 Step C — wait group
// // UPDATED: Worker function with sync.WaitGroup
// func prepareOrder(orderID string, prepTime time.Duration, resultChan chan<- OrderResult, wg *sync.WaitGroup) {
// 	defer wg.Done() // Signal that this goroutine is complete before exiting

// 	startTime := time.Now()
// 	time.Sleep(prepTime)

//		elapsed := time.Since(startTime)
//		resultChan <- OrderResult{
//			OrderID:  orderID,
//			Status:   "READY",
//			PrepTime: elapsed,
//		}
//	}
//
// 5.4
// UPDATED: Worker function now accepts the full Order struct
func prepareOrder(order Order, prepTime time.Duration, resultChan chan<- OrderResult, wg *sync.WaitGroup) {
	defer wg.Done()

	startTime := time.Now()
	time.Sleep(prepTime)

	elapsed := time.Since(startTime)
	resultChan <- OrderResult{
		OrderID:  order.OrderID, // Extract the ID directly from the struct
		Status:   "READY",
		PrepTime: elapsed,
	}
}

func main() {
	//==================LAB2=======================//

	// 5.1 Step A — Concurrent Order Preparation 5.1.1
	// doneChan := make(chan string, 3)

	// // Run three orders sequentially (one after the other)
	// // 5.2 Step B — Sequential Order Preparation
	// prepareOrderSequential("ORD-1001", 3*time.Second) //[cite: 1]
	// prepareOrderSequential("ORD-1002", 1*time.Second)
	// prepareOrderSequential("ORD-1003", 2*time.Second)
	// Create a channel that accepts the new OrderResult struct
	//5.2 Step B — Concurrent Order Preparation with Channels
	// resultChan := make(chan OrderResult, 3)

	// // Run concurrent orders (using IDs from your example)
	// go prepareOrder("101", 500*time.Millisecond, resultChan)
	// go prepareOrder("102", 200*time.Millisecond, resultChan)
	// go prepareOrder("103", 300*time.Millisecond, resultChan)

	// fmt.Println("=== Completed Orders ===")

	// // Wait and listen for the results
	// for i := 0; i < 3; i++ {
	// 	result := <-resultChan
	// 	fmt.Printf("Order #%s -- %s -- %v\n", result.OrderID, result.Status, result.PrepTime)
	// }
	// 5.3 Step C — wait group

	resultChan := make(chan OrderResult, 3)
	var wg sync.WaitGroup // Initialize the WaitGroup

	// We have 3 orders, so add 3 to the WaitGroup counter
	wg.Add(3)
	// 1. Thông báo worker đang chạy (đặt ngay sau khi khởi chạy các goroutine)
	fmt.Println("workers are running...")
	// Pass the memory address of wg (&wg) to each worker
	// go prepareOrder("101", 500*time.Millisecond, resultChan, &wg)
	// go prepareOrder("102", 200*time.Millisecond, resultChan, &wg)
	// go prepareOrder("103", 300*time.Millisecond, resultChan, &wg)
	// Assuming myOrder is already created
	go prepareOrder(myOrder, 500*time.Millisecond, resultChan, &wg)
	// 2. Thông báo tất cả worker đã hoàn thành (vòng lặp channel đã kết thúc)

	// Wait for all workers to finish in the background, then close the channel
	go func() {
		wg.Wait()
		close(resultChan)
		fmt.Println("all workers complete")
	}()

	fmt.Println("=== Completed Orders ===")

	// Read continuously until the channel is closed
	for result := range resultChan {
		fmt.Printf("Order #%s -- %s -- %v\n", result.OrderID, result.Status, result.PrepTime)
	}

}
