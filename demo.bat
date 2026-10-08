@echo off
title Order Service - gRPC demo
cd /d "C:\Users\atuav\Desktop\Proyecto Sistemas\order-svc"
set PATH=%PATH%;C:\Users\atuav\AppData\Local\Programs\grpcurl
mode con lines=48 cols=150

echo C:\order-svc^> grpcurl -plaintext -import-path . -proto pb/orders.proto list orders.OrderService
grpcurl -plaintext -import-path . -proto pb/orders.proto list orders.OrderService
echo.
echo C:\order-svc^> grpcurl -plaintext -import-path . -proto pb/orders.proto -d {"order_id": "ORD-1001"} localhost:50051 orders.OrderService/GetOrder
grpcurl -plaintext -import-path . -proto pb/orders.proto -d "{\"order_id\": \"ORD-1001\"}" localhost:50051 orders.OrderService/GetOrder
echo.
pause
