package main

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

// =============================================================================
// Product & ProductFactory
// =============================================================================

type Product struct {
	SKU   int
	Name  string
	Price float64
}

func CreateProduct(sku int) *Product {
	// In reality this comes from DB
	var name string
	var price float64

	switch sku {
	case 101:
		name, price = "Apple", 20
	case 102:
		name, price = "Banana", 10
	case 103:
		name, price = "Chocolate", 50
	case 201:
		name, price = "T-Shirt", 500
	case 202:
		name, price = "Jeans", 1000
	default:
		name, price = fmt.Sprintf("Item%d", sku), 100
	}
	return &Product{SKU: sku, Name: name, Price: price}
}

// =============================================================================
// InventoryStore (Interface) & DbInventoryStore
// =============================================================================

type InventoryStore interface {
	AddProduct(prod *Product, qty int)
	RemoveProduct(sku, qty int)
	CheckStock(sku int) int
	ListAvailableProducts() []*Product
}

// DbInventoryStore is an in-memory store (simulates DB).
// Uses a sync.RWMutex to be safe for concurrent use.
type DbInventoryStore struct {
	mu       sync.RWMutex
	stock    map[int]int      // SKU -> quantity
	products map[int]*Product // SKU -> Product
}

func NewDbInventoryStore() *DbInventoryStore {
	return &DbInventoryStore{
		stock:    make(map[int]int),
		products: make(map[int]*Product),
	}
}

func (s *DbInventoryStore) AddProduct(prod *Product, qty int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[prod.SKU]; !exists {
		s.products[prod.SKU] = prod
	}
	s.stock[prod.SKU] += qty
}

func (s *DbInventoryStore) RemoveProduct(sku, qty int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, exists := s.stock[sku]
	if !exists {
		return
	}
	remaining := current - qty
	if remaining > 0 {
		s.stock[sku] = remaining
	} else {
		delete(s.stock, sku)
		delete(s.products, sku)
	}
}

func (s *DbInventoryStore) CheckStock(sku int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stock[sku]
}

func (s *DbInventoryStore) ListAvailableProducts() []*Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var available []*Product
	for sku, qty := range s.stock {
		if qty > 0 {
			if p, ok := s.products[sku]; ok {
				available = append(available, p)
			}
		}
	}
	return available
}

// =============================================================================
// InventoryManager
// =============================================================================

type InventoryManager struct {
	store InventoryStore
}

func NewInventoryManager(store InventoryStore) *InventoryManager {
	return &InventoryManager{store: store}
}

func (m *InventoryManager) AddStock(sku, qty int) {
	prod := CreateProduct(sku)
	m.store.AddProduct(prod, qty)
	fmt.Printf("[InventoryManager] Added SKU %d Qty %d\n", sku, qty)
}

func (m *InventoryManager) RemoveStock(sku, qty int) {
	m.store.RemoveProduct(sku, qty)
}

func (m *InventoryManager) CheckStock(sku int) int {
	return m.store.CheckStock(sku)
}

func (m *InventoryManager) GetAvailableProducts() []*Product {
	return m.store.ListAvailableProducts()
}

// =============================================================================
// Replenishment Strategy (Strategy Pattern)
// =============================================================================

type ReplenishStrategy interface {
	Replenish(manager *InventoryManager, itemsToReplenish map[int]int)
}

// ThresholdReplenishStrategy replenishes if stock < threshold
type ThresholdReplenishStrategy struct {
	Threshold int
}

func (t *ThresholdReplenishStrategy) Replenish(manager *InventoryManager, itemsToReplenish map[int]int) {
	fmt.Println("[ThresholdReplenish] Checking threshold...")
	for sku, qtyToAdd := range itemsToReplenish {
		current := manager.CheckStock(sku)
		if current < t.Threshold {
			manager.AddStock(sku, qtyToAdd)
			fmt.Printf("  -> SKU %d was %d, replenished by %d\n", sku, current, qtyToAdd)
		}
	}
}

// WeeklyReplenishStrategy is a stub for weekly replenishment
type WeeklyReplenishStrategy struct{}

func (w *WeeklyReplenishStrategy) Replenish(manager *InventoryManager, itemsToReplenish map[int]int) {
	fmt.Println("[WeeklyReplenish] Weekly replenishment triggered for inventory.")
}

// =============================================================================
// DarkStore
// =============================================================================

type DarkStore struct {
	Name             string
	X, Y             float64
	inventoryManager *InventoryManager
	replenishStrategy ReplenishStrategy
}

func NewDarkStore(name string, x, y float64) *DarkStore {
	return &DarkStore{
		Name:             name,
		X:                x,
		Y:                y,
		inventoryManager: NewInventoryManager(NewDbInventoryStore()),
	}
}

func (ds *DarkStore) SetReplenishStrategy(strategy ReplenishStrategy) {
	ds.replenishStrategy = strategy
}

func (ds *DarkStore) DistanceTo(ux, uy float64) float64 {
	dx := ds.X - ux
	dy := ds.Y - uy
	return math.Sqrt(dx*dx + dy*dy)
}

func (ds *DarkStore) RunReplenishment(itemsToReplenish map[int]int) {
	if ds.replenishStrategy != nil {
		ds.replenishStrategy.Replenish(ds.inventoryManager, itemsToReplenish)
	}
}

func (ds *DarkStore) GetAllProducts() []*Product {
	return ds.inventoryManager.GetAvailableProducts()
}

func (ds *DarkStore) CheckStock(sku int) int {
	return ds.inventoryManager.CheckStock(sku)
}

func (ds *DarkStore) RemoveStock(sku, qty int) {
	ds.inventoryManager.RemoveStock(sku, qty)
}

func (ds *DarkStore) AddStock(sku, qty int) {
	ds.inventoryManager.AddStock(sku, qty)
}

// =============================================================================
// DarkStoreManager (Singleton)
// =============================================================================

type DarkStoreManager struct {
	darkStores []*DarkStore
}

var (
	darkStoreManagerInstance *DarkStoreManager
	darkStoreManagerOnce     sync.Once
)

func GetDarkStoreManager() *DarkStoreManager {
	darkStoreManagerOnce.Do(func() {
		darkStoreManagerInstance = &DarkStoreManager{}
	})
	return darkStoreManagerInstance
}

func (m *DarkStoreManager) RegisterDarkStore(ds *DarkStore) {
	m.darkStores = append(m.darkStores, ds)
}

func (m *DarkStoreManager) GetNearbyDarkStores(ux, uy, maxDistance float64) []*DarkStore {
	type distStore struct {
		dist  float64
		store *DarkStore
	}
	var distList []distStore

	for _, ds := range m.darkStores {
		d := ds.DistanceTo(ux, uy)
		if d <= maxDistance {
			distList = append(distList, distStore{d, ds})
		}
	}

	sort.Slice(distList, func(i, j int) bool {
		return distList[i].dist < distList[j].dist
	})

	result := make([]*DarkStore, 0, len(distList))
	for _, item := range distList {
		result = append(result, item.store)
	}
	return result
}

// =============================================================================
// User & Cart
// =============================================================================

type CartItem struct {
	Product *Product
	Qty     int
}

type Cart struct {
	Items []CartItem
}

func (c *Cart) AddItem(sku, qty int) {
	prod := CreateProduct(sku)
	c.Items = append(c.Items, CartItem{Product: prod, Qty: qty})
	fmt.Printf("[Cart] Added SKU %d (%s) x%d\n", sku, prod.Name, qty)
}

func (c *Cart) GetTotal() float64 {
	var total float64
	for _, item := range c.Items {
		total += item.Product.Price * float64(item.Qty)
	}
	return total
}

type User struct {
	Name string
	X, Y float64
	cart *Cart
}

func NewUser(name string, x, y float64) *User {
	return &User{Name: name, X: x, Y: y, cart: &Cart{}}
}

func (u *User) GetCart() *Cart {
	return u.cart
}

// =============================================================================
// DeliveryPartner
// =============================================================================

type DeliveryPartner struct {
	Name string
}

// =============================================================================
// Order & OrderManager (Singleton)
// =============================================================================

type OrderItem struct {
	Product *Product
	Qty     int
}

type Order struct {
	OrderID     int
	User        *User
	Items       []OrderItem
	Partners    []*DeliveryPartner
	TotalAmount float64
}

type OrderManager struct {
	mu      sync.Mutex
	orders  []*Order
	nextID  int
}

var (
	orderManagerInstance *OrderManager
	orderManagerOnce     sync.Once
)

func GetOrderManager() *OrderManager {
	orderManagerOnce.Do(func() {
		orderManagerInstance = &OrderManager{nextID: 1}
	})
	return orderManagerInstance
}

func (om *OrderManager) PlaceOrder(user *User, cart *Cart) {
	fmt.Printf("\n[OrderManager] Placing Order for: %s\n", user.Name)

	requestedItems := cart.Items

	// 1) Find nearby dark stores within 5 KM
	const maxDist = 5.0
	nearbyStores := GetDarkStoreManager().GetNearbyDarkStores(user.X, user.Y, maxDist)

	if len(nearbyStores) == 0 {
		fmt.Println("  No dark stores within 5 KM. Cannot fulfill order.")
		return
	}

	// 2) Check if closest store has ALL items
	firstStore := nearbyStores[0]
	allInFirst := true
	for _, item := range requestedItems {
		if firstStore.CheckStock(item.Product.SKU) < item.Qty {
			allInFirst = false
			break
		}
	}

	om.mu.Lock()
	order := &Order{OrderID: om.nextID, User: user}
	om.nextID++
	om.mu.Unlock()

	if allInFirst {
		// Single store fulfillment - one delivery partner
		fmt.Printf("  All items at: %s\n", firstStore.Name)
		for _, item := range requestedItems {
			firstStore.RemoveStock(item.Product.SKU, item.Qty)
			order.Items = append(order.Items, OrderItem{Product: item.Product, Qty: item.Qty})
		}
		order.TotalAmount = cart.GetTotal()
		order.Partners = append(order.Partners, &DeliveryPartner{Name: "Partner1"})
		fmt.Println("  Assigned Delivery Partner: Partner1")

	} else {
		// Multi-store fulfillment - multiple delivery partners
		fmt.Println("  Splitting order across stores...")

		// Build a remaining-qty map
		remaining := make(map[int]int)
		for _, item := range requestedItems {
			remaining[item.Product.SKU] = item.Qty
		}

		partnerID := 1
		for _, store := range nearbyStores {
			if len(remaining) == 0 {
				break
			}
			fmt.Printf("   Checking: %s\n", store.Name)

			var fulfilledFromThisStore []int
			for sku, qtyNeeded := range remaining {
				available := store.CheckStock(sku)
				if available <= 0 {
					continue
				}
				taken := min(available, qtyNeeded)
				store.RemoveStock(sku, taken)
				fmt.Printf("     %s supplies SKU %d x%d\n", store.Name, sku, taken)
				order.Items = append(order.Items, OrderItem{
					Product: CreateProduct(sku),
					Qty:     taken,
				})
				if qtyNeeded > taken {
					remaining[sku] = qtyNeeded - taken
				} else {
					fulfilledFromThisStore = append(fulfilledFromThisStore, sku)
				}
			}

			for _, sku := range fulfilledFromThisStore {
				delete(remaining, sku)
			}

			if len(fulfilledFromThisStore) > 0 {
				pName := fmt.Sprintf("Partner%d", partnerID)
				partnerID++
				order.Partners = append(order.Partners, &DeliveryPartner{Name: pName})
				fmt.Printf("     Assigned: %s for %s\n", pName, store.Name)
			}
		}

		if len(remaining) > 0 {
			fmt.Println("  Could not fulfill:")
			for sku, qty := range remaining {
				fmt.Printf("    SKU %d x%d\n", sku, qty)
			}
		}

		var sum float64
		for _, item := range order.Items {
			sum += item.Product.Price * float64(item.Qty)
		}
		order.TotalAmount = sum
	}

	// Print Order Summary
	fmt.Printf("\n[OrderManager] Order #%d Summary:\n", order.OrderID)
	fmt.Printf("  User: %s\n  Items:\n", user.Name)
	for _, item := range order.Items {
		fmt.Printf("    SKU %d (%s) x%d @ ₹%.2f\n",
			item.Product.SKU, item.Product.Name, item.Qty, item.Product.Price)
	}
	fmt.Printf("  Total: ₹%.2f\n  Partners:\n", order.TotalAmount)
	for _, dp := range order.Partners {
		fmt.Printf("    %s\n", dp.Name)
	}
	fmt.Println()

	om.mu.Lock()
	om.orders = append(om.orders, order)
	om.mu.Unlock()
}

func (om *OrderManager) GetAllOrders() []*Order {
	om.mu.Lock()
	defer om.mu.Unlock()
	return om.orders
}

// =============================================================================
// ZeptoHelper
// =============================================================================

func ShowAllItems(user *User) {
	fmt.Printf("\n[Zepto] All Available products within 5 KM for %s:\n", user.Name)
	nearbyStores := GetDarkStoreManager().GetNearbyDarkStores(user.X, user.Y, 5.0)

	skuToPrice := make(map[int]float64)
	skuToName := make(map[int]string)

	for _, ds := range nearbyStores {
		for _, product := range ds.GetAllProducts() {
			if _, seen := skuToPrice[product.SKU]; !seen {
				skuToPrice[product.SKU] = product.Price
				skuToName[product.SKU] = product.Name
			}
		}
	}

	for sku, price := range skuToPrice {
		fmt.Printf("  SKU %d - %s @ ₹%.2f\n", sku, skuToName[sku], price)
	}
}

func Initialize() {
	dsManager := GetDarkStoreManager()

	// DarkStore A
	storeA := NewDarkStore("DarkStoreA", 0.0, 0.0)
	storeA.SetReplenishStrategy(&ThresholdReplenishStrategy{Threshold: 3})
	fmt.Println("\nAdding stocks in DarkStoreA....")
	storeA.AddStock(101, 5)
	storeA.AddStock(102, 2)

	// DarkStore B
	storeB := NewDarkStore("DarkStoreB", 4.0, 1.0)
	storeB.SetReplenishStrategy(&ThresholdReplenishStrategy{Threshold: 3})
	fmt.Println("\nAdding stocks in DarkStoreB....")
	storeB.AddStock(101, 3)
	storeB.AddStock(103, 10)

	// DarkStore C
	storeC := NewDarkStore("DarkStoreC", 2.0, 3.0)
	storeC.SetReplenishStrategy(&ThresholdReplenishStrategy{Threshold: 3})
	fmt.Println("\nAdding stocks in DarkStoreC....")
	storeC.AddStock(102, 5)
	storeC.AddStock(201, 7)

	dsManager.RegisterDarkStore(storeA)
	dsManager.RegisterDarkStore(storeB)
	dsManager.RegisterDarkStore(storeC)
}

// =============================================================================
// min helper (Go 1.20 has built-in min, this ensures compatibility)
// =============================================================================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// Main
// =============================================================================

func main() {
	// 1) Initialize dark stores
	Initialize()

	// 2) A user comes on the platform
	user := NewUser("Aditya", 1.0, 1.0)
	fmt.Printf("\nUser with name %s comes on platform\n", user.Name)

	// 3) Show all available items nearby
	ShowAllItems(user)

	// 4) User adds items to cart
	fmt.Println("\nAdding items to cart")
	cart := user.GetCart()
	cart.AddItem(101, 4)
	cart.AddItem(102, 3)
	cart.AddItem(103, 2)

	// 5) Place Order
	GetOrderManager().PlaceOrder(user, cart)

	fmt.Println("\n=== Demo Complete ===")
}
