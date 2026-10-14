package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	StatusWaiting   = "WAITING"
	StatusPreparing = "PREPARING"
	StatusCancelled = "CANCELLED"
	StatusReady     = "READY"
)

type Order struct {
	ID       int
	Items    string
	Duration time.Duration
	Status   string
	mu       *sync.Mutex
}

type OrderResult struct {
	OrderID  int
	Status   string
	Duration time.Duration
}

func CancelOrder(order *Order) bool {
	order.mu.Lock()
	defer order.mu.Unlock()

	// An order can only be cancelled if it is still waiting
	if order.Status == StatusWaiting {
		order.Status = StatusCancelled
		return true
	}
	return false
}

func prepareOrder(order *Order, results chan<- OrderResult, slots chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// Wait for an available kitchen slot
	slots <- struct{}{}

	// Check if the order was cancelled while waiting
	order.mu.Lock()
	if order.Status == StatusCancelled {
		order.mu.Unlock()
		<-slots // Release the slot immediately

		// Report the cancellation back to the main thread
		results <- OrderResult{OrderID: order.ID, Status: StatusCancelled, Duration: 0}
		return // Order is never delivered to the active kitchen worker
	}
	// If not cancelled, mark as preparing
	order.Status = StatusPreparing
	order.mu.Unlock()

	fmt.Printf("[START] Order #%d is cooking...\n", order.ID)
	time.Sleep(order.Duration)

	order.mu.Lock()
	order.Status = StatusReady
	order.mu.Unlock()

	results <- OrderResult{
		OrderID:  order.ID,
		Status:   StatusReady,
		Duration: order.Duration,
	}

	<-slots // Release slot
}

func runCancelTest(testName string, queue []*Order, cancelFunc func()) {
	fmt.Printf("\n%s\n", testName)

	results := make(chan OrderResult, len(queue))
	slots := make(chan struct{}, 2) // Kitchen capacity = 2
	var wg sync.WaitGroup

	start := time.Now()

	// Dispatch orders
	for _, order := range queue {
		wg.Add(1)
		go prepareOrder(order, results, slots, &wg)
	}

	// Execute the specific cancellation logic for this test
	if cancelFunc != nil {
		cancelFunc()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("Report")
	for res := range results {
		fmt.Printf("[DONE] Order #%d %s\n", res.OrderID, res.Status)
	}

	fmt.Printf("Total execution time: %v\n", time.Since(start))
}

func main() {
	// Test 1: Normal execution (No cancellations)
	runCancelTest("Test 1: Normal Execution", []*Order{
		{ID: 101, Items: "Pho", Duration: 200 * time.Millisecond, Status: StatusWaiting, mu: &sync.Mutex{}},
	}, nil)

	// Test 2: Cancel before preparation
	order2 := &Order{ID: 102, Items: "Burger", Duration: 300 * time.Millisecond, Status: StatusWaiting, mu: &sync.Mutex{}}
	runCancelTest("Test 2: Cancel Before Preparation", []*Order{order2}, func() {
		if CancelOrder(order2) {
			fmt.Println("-> Successfully cancelled Order #102")
		}
	})

	// Test 3: Cancel an active order (Should Fail)
	order3 := &Order{ID: 103, Items: "Steak", Duration: 400 * time.Millisecond, Status: StatusWaiting, mu: &sync.Mutex{}}
	runCancelTest("Test 3: Cancel Active Order", []*Order{order3}, func() {
		time.Sleep(100 * time.Millisecond) // Let it enter PREPARING state
		if !CancelOrder(order3) {
			fmt.Println("-> Failed to cancel Order #103 (Already preparing)")
		}
	})

	// Test 4: Cancel an already completed order (Should Fail)
	order4 := &Order{ID: 104, Items: "Coffee", Duration: 100 * time.Millisecond, Status: StatusWaiting, mu: &sync.Mutex{}}
	runCancelTest("Test 4: Cancel Completed Order", []*Order{order4}, func() {
		time.Sleep(200 * time.Millisecond) // Let it finish completely
		if !CancelOrder(order4) {
			fmt.Println("-> Failed to cancel Order #104 (Already completed)")
		}
	})

	// Test 5: Complex queue with multiple waiting cancellations
	o601 := &Order{ID: 601, Items: "Pho", Duration: 500 * time.Millisecond, Status: StatusWaiting, mu: &sync.Mutex{}}
	o602 := &Order{ID: 602, Items: "Banh Mi", Duration: 500 * time.Millisecond, Status: StatusWaiting, mu: &sync.Mutex{}}
	o603 := &Order{ID: 603, Items: "Coffee", Duration: 200 * time.Millisecond, Status: StatusWaiting, mu: &sync.Mutex{}}
	o604 := &Order{ID: 604, Items: "Burger", Duration: 200 * time.Millisecond, Status: StatusWaiting, mu: &sync.Mutex{}}
	o605 := &Order{ID: 605, Items: "Tea", Duration: 200 * time.Millisecond, Status: StatusWaiting, mu: &sync.Mutex{}}
	
	runCancelTest("TEST 5: Queue with Orders #601 - #605", []*Order{o601, o602, o603, o604, o605}, func() {
		// Wait 50ms so workers take 601 and 602 into the 2 kitchen slots
		time.Sleep(50 * time.Millisecond)

		// Try to cancel 601 (Should fail because it is already cooking)
		if !CancelOrder(o601) {
			fmt.Println("-> Cannot cancel Order #601: already in cooking process!")
		}

		// Cancel 603 and 605 (Should succeed because slots are full and workers haven't taken them yet)
		if CancelOrder(o603) {
			fmt.Println("-> Successfully canceled waiting Order #603")
		}
		if CancelOrder(o605) {
			fmt.Println("-> Successfully canceled waiting Order #605")
		}
	})
}
