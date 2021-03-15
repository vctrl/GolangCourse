package exchange

import (
	context "context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	grpc "google.golang.org/grpc"
)

const listenAddr = "127.0.0.1:8082"
const fileName = "../../SPFB.RTS_190517_190517.txt"

type TradingSourceMock struct {
}

type Case struct {
	deals []*Deal
	order *Deal
}

// deals := []*Deal{
// 	{Ticker: "TEST", Amount: 1, Price: 10},
// 	{Ticker: "TEST", Amount: 1, Price: 20},
// 	{Ticker: "TEST", Amount: 1, Price: 30},
// 	{Ticker: "TEST", Amount: 1, Price: 40},
// 	{Ticker: "TEST", Amount: 1, Price: 50},
// }
// _, err = client.Create(context.Background(), &Deal{Ticker: "TEST", BrokerID: 1, Price: float32(25), Amount: -1})

// partial
deals := []*Deal{
	{Ticker: "TEST", Amount: 1, Price: 10},
	{Ticker: "TEST", Amount: 1, Price: 20},
	{Ticker: "TEST", Amount: 1, Price: 10},
	{Ticker: "TEST", Amount: 1, Price: 20},
	{Ticker: "TEST", Amount: 1, Price: 10},
	{Ticker: "TEST", Amount: 1, Price: 20},
}
_, err = client.Create(context.Background(), &Deal{Ticker: "TEST", BrokerID: 1, Price: float32(25), Amount: -3})


func (*TradingSourceMock) StartTrading(input io.Reader, es *ExchangeServerImpl) {
	deals := []*Deal{
		{Ticker: "TEST", Amount: 1, Price: 10},
		{Ticker: "TEST", Amount: 1, Price: 20},
		{Ticker: "TEST", Amount: 1, Price: 30},
		{Ticker: "TEST", Amount: 1, Price: 40},
		{Ticker: "TEST", Amount: 1, Price: 50},
	}

	// ждём подписок
	time.Sleep(time.Millisecond * 100)
	for _, d := range deals {
		es.Deals <- d
	}
}

// 1 заявка на продажу, цена подходит снизу вверх, объёма хватает: успех
// 2 заявка на покупку, цена подходит сверху вниз, объёма хватает: успех
// 3 заявка на продажу, цена подходит снизу вверх, сначала объёма не хватает потом хватает: успех, флаг partial
// 4 заявка на покупку, цена подходит сверху вниз, сначала объёма не хватает потом хватает: успех, флаг partial

// проверка того, что сначала приоритет на цену
// проверка того, что есть приоритет в порядке добавления

// 7 заявка, отмена, ожидание: ничего не произошло

// 8 статистика, два брокера, каждую секунду получают одинаковую статистику

func TestOrderResults(t *testing.T) {
	grpcConn, err := grpc.Dial(
		listenAddr,
		grpc.WithInsecure(),
	)

	if err != nil {
		log.Fatalf("cant connect to grpc: %v", err)
	}
	defer grpcConn.Close()

	f, err := os.Open("../../SPFB.RTS_190517_190517.txt")
	defer f.Close()
	if err != nil {
		log.Fatalf("error opening file, %+v", err)
	}

	go StartExchangeService(context.Background(), listenAddr, &TradingSourceMock{}, f)
	time.Sleep(time.Millisecond * 10)

	client := NewExchangeClient(grpcConn)
	res1, err := client.Results(context.Background(), &BrokerID{ID: 1})
	_, err = client.Create(context.Background(), &Deal{Ticker: "TEST", BrokerID: 1, Price: float32(25), Amount: -1})

	res2, err := client.Results(context.Background(), &BrokerID{ID: 2})
	_, err = client.Create(context.Background(), &Deal{Ticker: "TEST", BrokerID: 2, Price: float32(35), Amount: -1})

	// expectedDeal1 := &Deal{}
	// expectedDeal2 := &Deal{}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	for i := 0; i < 1; i++ {
		go func() {
			defer wg.Done()
			d1, err := res1.Recv()
			fmt.Println(d1)
			if err != nil {
				// t.Fatalf("unexpected error: %v", err)
			}

			d2, err := res2.Recv()
			if err != nil {
				// t.Fatalf("unexpected error: %v", err)
			}
			fmt.Println(d2)

		}()
	}

	wg.Wait()
}
