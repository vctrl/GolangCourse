package main

import (
	"context"
	"trading/grpc/exchange"
)

func main() {
	exchange.StartExchangeService(context.Background(), "127.0.0.1:8082")
}
