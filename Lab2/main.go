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
	Priority bool
	Status   string
}

type OrderResult struct {
	OrderID  int
	Status   string
	Duration time.Duration
}

// kitchenWorker ONLY receives valid orders that are already marked PREPARING.
// Canceled orders are NEVER delivered to this function.
func kitchenWorker(order *Order, results chan<- OrderResult, slots chan struct{}, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	fmt.Printf("[START] Order #%d (%s) is cooking...\n", order.ID, order.Items)
	time.Sleep(order.Duration)

	mu.Lock()
	order.Status = StatusReady
	mu.Unlock()

	results <- OrderResult{
		OrderID:  order.ID,
		Status:   StatusReady,
		Duration: order.Duration,
	}

	<-slots // Release the kitchen slot so the queue can send the next order
}

func runCancelTest(testName string, queue []Order, cancelIDs []int, cancelDelay time.Duration) {
	fmt.Printf("\n========================================\n")
	fmt.Printf("HOMEWORK: %s\n", testName)
	fmt.Printf("========================================\n")

	if len(queue) == 0 {
		fmt.Println("Queue is empty.")
		return
	}

	var mu sync.Mutex
	orderMap := make(map[int]*Order)
	for i := range queue {
		queue[i].Status = StatusWaiting
		orderMap[queue[i].ID] = &queue[i]
	}

	results := make(chan OrderResult, len(queue))
	slots := make(chan struct{}, 2) // Max 2 active orders in the kitchen
	var wg sync.WaitGroup

	start := time.Now()

	// Helper function to process cancellations
	doCancel := func() {
		for _, id := range cancelIDs {
			mu.Lock()
			o, exists := orderMap[id]
			if exists {
				if o.Status == StatusWaiting {
					o.Status = StatusCancelled
					fmt.Printf("[CANCEL] Order #%d CANCELED while waiting in queue\n", id)
				} else {
					fmt.Printf("[REJECT] Order #%d CANNOT be canceled (already %s)\n", id, o.Status)
				}
			}
			mu.Unlock()
		}
	}

	// If delay is 0 (like Test C6), user cancels immediately while all orders are in the queue
	if cancelDelay == 0 {
		doCancel()
	} else {
		// Otherwise, run a timer in the background to cancel at cancelDelay
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(cancelDelay)
			doCancel()
		}()
	}

	// QUEUE DISPATCHER: Wait for a kitchen slot BEFORE pulling the next order from the queue
	go func() {
		for i := range queue {
			// 1. Wait until one of the 2 kitchen slots is free
			slots <- struct{}{}

			mu.Lock()
			o := &queue[i]

			// 2. Check if the order was canceled while sitting in the queue
			if o.Status == StatusCancelled {
				mu.Unlock()
				<-slots // Free the slot immediately
				// Record the canceled status in the report, NEVER send to kitchenWorker
				results <- OrderResult{OrderID: o.ID, Status: StatusCancelled, Duration: 0}
				continue
			}

			// 3. Not canceled! Lock it into PREPARING state and deliver to kitchenWorker
			o.Status = StatusPreparing
			mu.Unlock()

			wg.Add(1)
			go kitchenWorker(o, results, slots, &wg, &mu)
		}
	}()

	// Wait a tiny bit for the dispatcher to start, then wait for all work to finish
	time.Sleep(20 * time.Millisecond)
	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("--- Report ---")
	for res := range results {
		fmt.Printf("[DONE] Order #%d %s (Took: %v)\n", res.OrderID, res.Status, res.Duration)
	}
	fmt.Printf("Total execution time: %v\n", time.Since(start))
}

func main() {
	// Test C1: Cancel two waiting orders
	runCancelTest("C1 - Cancel Two Waiting Orders", []Order{
		{ID: 701, Items: "Burger", Duration: 300 * time.Millisecond},
		{ID: 702, Items: "Fries", Duration: 300 * time.Millisecond},
		{ID: 703, Items: "Soup", Duration: 200 * time.Millisecond},
		{ID: 704, Items: "Drink", Duration: 200 * time.Millisecond},
	}, []int{703, 704}, 100*time.Millisecond)

	// Test C2: Cancel active orders (both rejected)
	runCancelTest("C2 - Cancel Active Orders Rejected", []Order{
		{ID: 711, Items: "Steak", Duration: 300 * time.Millisecond},
		{ID: 712, Items: "Soup", Duration: 300 * time.Millisecond},
		{ID: 713, Items: "Salad", Duration: 200 * time.Millisecond},
	}, []int{711, 712}, 100*time.Millisecond)

	// Test C3: Cancel some waiting orders, others still cook
	runCancelTest("C3 - Partial Cancel Among Many Waiting", []Order{
		{ID: 721, Items: "Dish 1", Duration: 300 * time.Millisecond},
		{ID: 722, Items: "Dish 2", Duration: 300 * time.Millisecond},
		{ID: 723, Items: "Dish 3", Duration: 200 * time.Millisecond},
		{ID: 724, Items: "Dish 4", Duration: 200 * time.Millisecond},
		{ID: 725, Items: "Dish 5", Duration: 200 * time.Millisecond},
	}, []int{723, 725}, 100*time.Millisecond)

	// Test C4: Cancel after the order already finished (rejected)
	runCancelTest("C4 - Cancel After Completion", []Order{
		{ID: 731, Items: "Tea", Duration: 100 * time.Millisecond},
	}, []int{731}, 300*time.Millisecond)

	// Test C5: Cancel the same waiting order twice (second is rejected)
	runCancelTest("C5 - Double Cancel", []Order{
		{ID: 741, Items: "Pho", Duration: 300 * time.Millisecond},
		{ID: 742, Items: "Coffee", Duration: 300 * time.Millisecond},
		{ID: 743, Items: "Banh Mi", Duration: 200 * time.Millisecond},
	}, []int{743, 743}, 100*time.Millisecond)

	// Test C6: Cancel all orders immediately
	runCancelTest("C6 - Cancel All Orders", []Order{
		{ID: 751, Items: "Dish 1", Duration: 300 * time.Millisecond},
		{ID: 752, Items: "Dish 2", Duration: 300 * time.Millisecond},
		{ID: 753, Items: "Dish 3", Duration: 200 * time.Millisecond},
		{ID: 754, Items: "Dish 4", Duration: 200 * time.Millisecond},
	}, []int{751, 752, 753, 754}, 0*time.Millisecond)
}
