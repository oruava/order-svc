package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/oruava/order-svc/pb"
)

// server implementa la interfaz generada por protoc
type server struct {
	pb.UnimplementedOrderServiceServer

	mu     sync.Mutex
	orders map[string]*pb.Order
}

// precios hardcodeados (proximamente: validados con Restaurant Service)
var prices = map[string]float64{
	"prod-1": 4990,
	"prod-2": 2990,
	"prod-3": 1990,
}

func (s *server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := timestamppb.New(time.Now())
	var total float64
	var items []*pb.OrderItem
	for _, it := range req.Items {
		price := prices[it.ProductId]
		subtotal := price * float64(it.Quantity)
		total += subtotal
		items = append(items, &pb.OrderItem{
			ProductId: it.ProductId,
			Quantity:  it.Quantity,
			UnitPrice: price,
			Subtotal:  subtotal,
		})
	}

	order := &pb.Order{
		OrderId:         fmt.Sprintf("ord-%d", time.Now().UnixNano()),
		ClientId:        req.ClientId,
		RestaurantId:    req.RestaurantId,
		Status:          "CREATED",
		DeliveryAddress: req.DeliveryAddress,
		TotalAmount:     total,
		Items:           items,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	s.orders[order.OrderId] = order

	log.Printf("pedido creado: %s (total=%.0f)", order.OrderId, total)
	return &pb.CreateOrderResponse{Order: order}, nil
}

func (s *server) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[req.OrderId]
	if !ok {
		// demo: pedido hardcodeado para poder probar sin crear antes
		order = &pb.Order{
			OrderId:         req.OrderId,
			ClientId:        "client-42",
			RestaurantId:    "rest-7",
			DeliveryId:      "",
			Status:          "PAID",
			DeliveryAddress: "Av. Alemania 0123, Temuco",
			TotalAmount:     9980,
			Items: []*pb.OrderItem{
				{ProductId: "prod-1", Quantity: 1, UnitPrice: 4990, Subtotal: 4990},
				{ProductId: "prod-2", Quantity: 1, UnitPrice: 2990, Subtotal: 2990},
				{ProductId: "prod-3", Quantity: 1, UnitPrice: 1990, Subtotal: 1990},
			},
			CreatedAt: timestamppb.New(time.Now().Add(-20 * time.Minute)),
			UpdatedAt: timestamppb.New(time.Now().Add(-5 * time.Minute)),
		}
	}
	return &pb.GetOrderResponse{Order: order}, nil
}

func (s *server) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[req.OrderId]
	if !ok {
		// demo: crea una entrada minima si el pedido no esta en memoria
		order = &pb.Order{OrderId: req.OrderId, Status: "CREATED"}
		s.orders[req.OrderId] = order
	}

	previous := order.Status
	order.Status = req.Status
	now := timestamppb.New(time.Now())
	order.UpdatedAt = now

	log.Printf("pedido %s: %s -> %s", order.OrderId, previous, req.Status)
	return &pb.UpdateOrderStatusResponse{
		OrderId:        order.OrderId,
		PreviousStatus: previous,
		NewStatus:      req.Status,
		UpdatedAt:      now,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("no se pudo escuchar: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterOrderServiceServer(srv, &server{orders: map[string]*pb.Order{}})

	log.Println("servidor gRPC escuchando en :50051")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("error al servir: %v", err)
	}
}
