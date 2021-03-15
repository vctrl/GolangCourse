package exchange

import (
	context "context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strconv"
	sync "sync"
	"time"

	grpc "google.golang.org/grpc"
)

var ToolsToBroadcast = []string{"SPFB.RTS"}
var StatSendIntervalInSeconds = 1

func StartExchangeService(ctx context.Context, listenAddr string, tradingSource TradingSource, f io.Reader) {
	server := grpc.NewServer()
	exch := NewExchangeServer(tradingSource)
	RegisterExchangeServer(server, exch)

	lis, _ := net.Listen("tcp", listenAddr)

	go func(l net.Listener, s *grpc.Server) {
		err := s.Serve(l)
		if err != nil {
			log.Fatal(err)
		}
	}(lis, server)

	go exch.TradingSource.StartTrading(f, exch)
	ticker := time.NewTicker(time.Second)
	stat := &Stat{}
	var prev *Deal
LOOP:
	for {
		select {
		case deal := <-exch.Deals:
			if prev != nil && exch.DOM.ExecuteBestDeal(deal.Price-prev.Price, deal) {
				exch.DOM.ExecutedDealEvents.Publish(deal)
			}
			prev = deal
		case <-ticker.C:
			for _, s := range stat.GetStat() {
				exch.StatEvents.Publish(s)
			}
		case <-ctx.Done():
			server.Stop()
			break LOOP
		}
	}
}

type ExchangeServerImpl struct {
	Deals         chan *Deal
	DOM           *DepthOfMarket
	StatEvents    *PubSubStats
	TradingSource TradingSource
}

func NewExchangeServer(ts TradingSource) *ExchangeServerImpl {
	dom := &DepthOfMarket{
		BuyOrders: &SortedPriceOrders{
			TickerOrders:  make(map[string][]*PriceOrders),
			DealIDByPrice: make(map[*DealID]int64),
			mu:            &sync.Mutex{},
		},
		SellOrders: &SortedPriceOrders{
			TickerOrders:  make(map[string][]*PriceOrders),
			DealIDByPrice: make(map[*DealID]int64),
			mu:            &sync.Mutex{},
		},
		ExecutedDealEvents: &PubSubDeals{subs: make(map[int64]chan *Deal), mu: &sync.Mutex{}},
	}
	return &ExchangeServerImpl{
		Deals: make(chan *Deal),
		StatEvents: &PubSubStats{
			subs: make(map[interface{}]chan *OHLCV),
			mu:   &sync.Mutex{}},
		DOM:           dom,
		TradingSource: ts,
	}
}

func (es *ExchangeServerImpl) Statistic(brokerID *BrokerID, out Exchange_StatisticServer) error {
	statCh := es.StatEvents.Subscribe(out)
	defer es.StatEvents.Unsubscribe()

	for s := range statCh {
		err := out.Send(s)
		if err != nil {
			return err
		}
	}

	return nil
}

// отправка на биржу заявки от брокера
func (es *ExchangeServerImpl) Create(ctx context.Context, d *Deal) (*DealID, error) {
	fmt.Println("creating order..")
	return es.DOM.AddDeal(d)
}

// отмена заявки
func (es *ExchangeServerImpl) Cancel(context.Context, *DealID) (*CancelResult, error) {
	// not implemented
	return nil, nil
}

// исполнение заявок от биржи к брокеру
// устанавливается 1 раз брокером и при исполнении какой-то заявки
func (es *ExchangeServerImpl) Results(brokerID *BrokerID, out Exchange_ResultsServer) error {
	executedDeals := es.DOM.ExecutedDealEvents.Subscribe(brokerID.GetID())
	defer es.DOM.ExecutedDealEvents.Unsubscribe()
	for d := range executedDeals {
		err := out.Send(d)
		if err != nil {
			return err
		}
		fmt.Printf("заявка %v исполнена!\n", d)
	}

	return nil
}

func (es *ExchangeServerImpl) mustEmbedUnimplementedExchangeServer() {

}

type TradingSource interface {
	StartTrading(input io.Reader, es *ExchangeServerImpl)
}

type FileTradingSource struct {
}

func (*FileTradingSource) StartTrading(input io.Reader, es *ExchangeServerImpl) {
	// parse header
	fmt.Println("starting reading deals from file..")
	r := csv.NewReader(input)

	fieldIDs := make(map[string]int)
	record, err := r.Read()
	if err == io.EOF {
		return
	}
	for value := range record {
		fieldIDs[record[value]] = value
		// fmt.Printf("%d, %s\n", value, fieldIDs[record[value]] )
	}

	deal := &Deal{}
	currTime := 100000
	nextSec := make(chan bool)

	go func() {
		for {
			record, err = r.Read()
			if err == io.EOF {
				break
			}

			deal.Ticker = record[fieldIDs["<TICKER>"]]
			t, err := strconv.Atoi(record[fieldIDs["<TIME>"]])
			if err == nil {
				deal.Time = int32(t)
			}
			price, err := strconv.ParseFloat(record[fieldIDs["<LAST>"]], 10)
			if err == nil {
				deal.Price = float32(price)
			}

			amount, err := strconv.ParseInt(record[fieldIDs["<VOL>"]], 10, 0)

			if err == nil {
				// possible overflow
				deal.Amount = int32(amount)
			}

			// look for placed orders

			if t != currTime {
				<-nextSec
				currTime++
			}

			es.Deals <- deal
			fmt.Printf("Прошла сделка %v\n", deal)
		}
	}()

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			nextSec <- true
		}
	}
}

type Stat struct {
	last  map[string]*Deal
	OHLCV map[string]*OHLCV
}

func newStat() *Stat {
	return &Stat{OHLCV: make(map[string]*OHLCV)}
}

func (s *Stat) GetStat() map[string]*OHLCV {
	for ticker, deal := range s.last {
		s.OHLCV[ticker].Close = deal.Price
	}

	return s.OHLCV
}

func (s *Stat) Update(d *Deal) (delta float32) {
	ts := s.OHLCV[d.Ticker]
	p := d.GetPrice()

	if ts.Open < 0 {
		ts.Open = p
	}
	if p > ts.High {
		ts.High = p
	}
	if p < ts.Low {
		ts.Low = p
	}
	ts.Volume += float32(d.GetAmount())
	s.last[d.Ticker] = d

	return d.Price - s.last[d.Ticker].Price
}

type PubSubStats struct {
	subs map[interface{}](chan *OHLCV)
	mu   *sync.Mutex
}

func (ps *PubSubStats) Subscribe(stream interface{}) chan *OHLCV {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.subs[stream] = make(chan *OHLCV)
	return ps.subs[stream]
}

func (ps *PubSubStats) Unsubscribe() {
	// todo
}

func (ps *PubSubStats) Publish(d *OHLCV) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, ch := range ps.subs {
		ch <- d
	}
}

type PubSubDeals struct {
	subs map[int64](chan *Deal) // int64 - BrokerID
	mu   *sync.Mutex
}

func (ps *PubSubDeals) Subscribe(brokerID int64) chan *Deal {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.subs[brokerID] = make(chan *Deal)
	return ps.subs[brokerID]
}

func (ps *PubSubDeals) Unsubscribe() {
	// todo
}

func (ps *PubSubDeals) Publish(d *Deal) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ch, ok := ps.subs[int64(d.BrokerID)]; ok {
		ch <- d
	}
}

// биржевой стакан
type DepthOfMarket struct {
	BuyOrders  *SortedPriceOrders
	SellOrders *SortedPriceOrders

	ExecutedDealEvents *PubSubDeals
}

// добавить заявку в стакан
func (dom *DepthOfMarket) AddDeal(d *Deal) (*DealID, error) {
	var dealID *DealID
	if d.Amount > 0 {
		dealID = dom.BuyOrders.Add(d)
	} else {
		dealID = dom.SellOrders.Add(d)
	}

	return dealID, nil
}

// выполняется на каждой сделке при изменении цены, определяем направление и ищем подходящую заявку
func (dom *DepthOfMarket) ExecuteBestDeal(delta float32, d *Deal) bool {
	if delta > 0 {
		if deal := dom.SellOrders.ExecuteMinInRange(d.Price-delta, d.Price, d.Ticker, d.Amount); deal != nil {
			dom.ExecutedDealEvents.Publish(deal)
			return true
		}
	}

	if delta < 0 {
		if deal := dom.BuyOrders.ExecuteMaxInRange(d.Price, d.Price-delta, d.Ticker, d.Amount); deal != nil {
			dom.ExecutedDealEvents.Publish(deal)
			return true
		}
	}

	return false
}

// очередь заявок по определённой цене
type PriceOrders struct {
	Price float32
	Deals *DealsQueue
}

type DealsQueue []*Deal

func NewDealsQueue() *DealsQueue {
	slice := make([]*Deal, 0, 100)
	dq := DealsQueue(slice)
	return &dq
}

func (q *DealsQueue) Enqueue(d *Deal) {
	*q = append(*q, d)
}

func (q *DealsQueue) Dequeue() *Deal {
	d := (*q)[0]
	*q = (*q)[1:]
	return d
}

func (q *DealsQueue) Peek() *Deal {
	return (*q)[0]
}

func (q *DealsQueue) Remove(id int64) {
	for i := 0; i < len(*q); i++ {
		if (*q)[i].ID == id {
			(*q)[i] = (*q)[len(*q)-1]
			(*q) = (*q)[:len(*q)-1]
		}
	}
}

type SortedPriceOrders struct {
	// Orders []*PriceOrders
	TickerOrders map[string][]*PriceOrders
	// мапа для удаления по ID сделки, можно будет сразу найти в какой очереди находится
	DealIDByPrice map[*DealID]int64
	lastID        int64

	mu *sync.Mutex
}

func (spo *SortedPriceOrders) Add(d *Deal) *DealID {
	spo.mu.Lock()
	defer spo.mu.Unlock()

	_, ok := spo.TickerOrders[d.Ticker]
	if !ok {
		spo.TickerOrders[d.Ticker] = make([]*PriceOrders, 0, 10)
	}
	for _, po := range spo.TickerOrders[d.Ticker] {
		if po.Price == d.Price {
			spo.lastID++
			d.ID = spo.lastID
			po.Deals.Enqueue(d)
			return &DealID{BrokerID: int64(d.BrokerID), ID: d.ID}
		}
	}

	new := &PriceOrders{Price: d.Price, Deals: NewDealsQueue()}
	spo.lastID++
	d.ID = spo.lastID
	new.Deals.Enqueue(d)
	spo.TickerOrders[d.Ticker] = append(spo.TickerOrders[d.Ticker], new)
	fmt.Println("ticker orders in add", spo.TickerOrders)

	sort.Slice(spo.TickerOrders[d.Ticker], func(i, j int) bool { return spo.TickerOrders[d.Ticker][i].Price < spo.TickerOrders[d.Ticker][j].Price })

	return &DealID{BrokerID: int64(d.BrokerID), ID: d.ID}
}

// исполнить ближайшую по цене заявку, первую в очереди
func (spo *SortedPriceOrders) ExecuteMinInRange(l, r float32, ticker string, amount int32) *Deal {
	spo.mu.Lock()
	defer spo.mu.Unlock()

	// брутфорс, сложность O(n)
	// можно бинарный поиск использовать

	fmt.Println("executing min in range", l, r)
	tickerOrders, ok := spo.TickerOrders[ticker]
	fmt.Println(tickerOrders)
	if !ok {
		return nil
	}

	for _, p := range tickerOrders {
		if p.Price >= l && p.Price <= r && len(*p.Deals) > 0 {
			deal := p.Deals.Peek()
			// partial execution
			if -deal.Amount < amount {
				deal.Amount += amount
				deal.Partial = true
				return deal
			}
			// amount is enough to process
			return p.Deals.Dequeue()
		}
	}

	return nil
}

func (spo *SortedPriceOrders) ExecuteMaxInRange(l, r float32, ticker string, amount int32) *Deal {
	spo.mu.Lock()
	defer spo.mu.Unlock()

	// брутфорс, сложность O(n)
	// можно бинарный поиск использовать

	tickerOrders, ok := spo.TickerOrders[ticker]
	if !ok {
		return nil
	}
	for i := len(tickerOrders) - 1; i >= 0; i-- {
		p := tickerOrders[i].Price
		deals := tickerOrders[i].Deals
		if p <= r && p >= l && len(*deals) > 0 {
			deal := deals.Peek()
			if deal.Amount < amount {
				// partial execution: leave in queue, but modify
				deal.Amount -= amount
				deal.Partial = true
				return deal
			}
			// amount is enough to process
			return deals.Dequeue()
		}
	}

	return nil
}
