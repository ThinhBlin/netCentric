package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Order defines the structure for all kitchen orders
type Order struct {
	ID       int
	Items    string
	Duration time.Duration
	Priority bool // Used in the ADAPT Phase
}

// OrderResult defines the result sent through the channel back to main
type OrderResult struct {
	OrderID  int
	Status   string
	Duration time.Duration
}

func prepareStandardOrder(order Order, results chan<- OrderResult, slots chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// Acquire kitchen slot (capacity limit = 2)
	slots <- struct{}{}

	fmt.Printf("[START] Order #%d is cooking...\n", order.ID)
	time.Sleep(order.Duration)

	results <- OrderResult{
		OrderID:  order.ID,
		Status:   "READY",
		Duration: order.Duration,
	}

	<-slots // Release slot
}

func runVerifyTest(testName string, queue []Order) {
	fmt.Printf("%s\n", testName)

	if len(queue) == 0 {
		fmt.Println("Queue is empty. System handled safely without hanging.")
		return
	}

	results := make(chan OrderResult, len(queue))
	slots := make(chan struct{}, 2)
	var wg sync.WaitGroup

	start := time.Now()

	for _, order := range queue {
		wg.Add(1)
		go prepareStandardOrder(order, results, slots, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("Report")
	for res := range results {
		fmt.Printf("[DONE] Order #%d %s (Took: %v)\n", res.OrderID, res.Status, res.Duration)
	}

	fmt.Printf("Total execution time: %v\n", time.Since(start))
}

// ---------------------------------------------------------------------
// ADAPT PHASE: PRIORITY DISPATCHER
// ---------------------------------------------------------------------
func runAdaptTest(testName string, queue []Order) {
	fmt.Printf("ADAPT: %s\n", testName)

	// Operational rule: Priority orders jump to the front of available preparation slots
	sort.SliceStable(queue, func(i, j int) bool {
		return queue[i].Priority && !queue[j].Priority
	})

	results := make(chan OrderResult, len(queue))
	slots := make(chan struct{}, 2)
	var wg sync.WaitGroup

	start := time.Now()

	for _, order := range queue {
		wg.Add(1)
		go func(o Order) {
			defer wg.Done()

			slots <- struct{}{}

			tag := "Normal"
			if o.Priority {
				tag = "PRIORITY"
			}
			fmt.Printf("[START] Order #%d (%s) is cooking...\n", o.ID, tag)
			time.Sleep(o.Duration)

			results <- OrderResult{
				OrderID:  o.ID,
				Status:   "READY",
				Duration: o.Duration,
			}

			<-slots
		}(order)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("Report")
	for res := range results {
		fmt.Printf("[DONE] Order #%d %s (Took: %v)\n", res.OrderID, res.Status, res.Duration)
	}

	fmt.Printf("Total execution time: %v\n", time.Since(start))
}

func main() {
	// Test 1: One order
	runVerifyTest("Test 1 - One Order", []Order{
		{ID: 101, Items: "Pho", Duration: 300 * time.Millisecond},
	})

	// Test 2: Five orders
	runVerifyTest("Test 2 - Five Orders Complete", []Order{
		{ID: 101, Items: "Pho", Duration: 200 * time.Millisecond},
		{ID: 102, Items: "Coffee", Duration: 200 * time.Millisecond},
		{ID: 103, Items: "Banh Mi", Duration: 200 * time.Millisecond},
		{ID: 104, Items: "Burger", Duration: 200 * time.Millisecond},
		{ID: 105, Items: "Tea", Duration: 200 * time.Millisecond},
	})

	// Test 3: Different durations (out of order finish)
	runVerifyTest("Test 3 - Different Durations", []Order{
		{ID: 201, Items: "Slow Roast", Duration: 500 * time.Millisecond},
		{ID: 202, Items: "Quick Drink", Duration: 150 * time.Millisecond},
		{ID: 203, Items: "Medium Snack", Duration: 300 * time.Millisecond},
	})

	// Test 4: Empty queue
	runVerifyTest("Test 4 - Empty Queue", []Order{})

	// Test 5: Two active orders (Capacity limit = 2)
	runVerifyTest("Test 5 - Exactly Two Active Orders", []Order{
		{ID: 301, Items: "Steak", Duration: 400 * time.Millisecond},
		{ID: 302, Items: "Soup", Duration: 400 * time.Millisecond},
	})

	// Test 6: More than two waiting orders proceed
	runVerifyTest("Test 6 - Waiting Orders Proceed Sequentially", []Order{
		{ID: 401, Items: "Dish 1", Duration: 250 * time.Millisecond},
		{ID: 402, Items: "Dish 2", Duration: 250 * time.Millisecond},
		{ID: 403, Items: "Waiting Dish 3", Duration: 150 * time.Millisecond},
		{ID: 404, Items: "Waiting Dish 4", Duration: 150 * time.Millisecond},
	})

	
	// ADAPT (Priority Orders Test)
	runAdaptTest("Requirement Change - Priority Customer Preference", []Order{
		{ID: 501, Items: "Normal Burger", Duration: 300 * time.Millisecond, Priority: false},
		{ID: 502, Items: "Normal Fries", Duration: 300 * time.Millisecond, Priority: false},
		{ID: 503, Items: "VIP Combo", Duration: 200 * time.Millisecond, Priority: true}, // PRIORITY
		{ID: 504, Items: "Normal Drink", Duration: 200 * time.Millisecond, Priority: false},
	})
}
