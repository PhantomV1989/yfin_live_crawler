package livedatacrawler

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/phantomv1989/yfin_live_crawler/yfinLive"
)

/*
This engine subscribes to Yahoo Finance live websocket data and pushes them to any Algo queues needing those info every minute
*/

var cookie = "B=5g3ji59etgsco&b=3&s=e6; PRF=t%3DAAPL%252BWU%252BGE%252BLSI%252BNIO%252BENB%252BXOM%252BWBC%252BACES%252BSGD%253DX%252B%255EGSPC%252BKDP%252BTOL%252BMAR%252BTTM%26qsp-fnncls-cue%3D1; ucs=lnct=1574577681; A1=d=AQABBL4TCl4CEASsi4RzaJaPCIq_HGotblsFEgEBAQFaC17yXq-0b2UB_SMAAAcImHHYXSpyDlg&S=AQAAApRpDpqOKLu8JCo18uflvPw; A3=d=AQABBL4TCl4CEASsi4RzaJaPCIq_HGotblsFEgEBAQFaC17yXq-0b2UB_SMAAAcImHHYXSpyDlg&S=AQAAApRpDpqOKLu8JCo18uflvPw; GUC=AQEBAQFeC1pe8kIdmwSW; A1S=d=AQABBL4TCl4CEASsi4RzaJaPCIq_HGotblsFEgEBAQFaC17yXq-0b2UB_SMAAAcImHHYXSpyDlg&S=AQAAApRpDpqOKLu8JCo18uflvPw&j=WORLD"

var client *http.Client
var socketMaxSubscriptionCount = 50
var debug = false

// OHLCVD for storing minute old data for processing
// Actual crawler to process this is not part of the project
type OHLCVD struct {
	Ohlc           [][]float64
	Vol            []int64
	Date           []int64
	Index          string
	Length         int
	FirstTime      int64
	LastTime       int64
	LastTimeString string
	Mut            sync.RWMutex
}

// LastMinState for storing last retrieved live data
type LastMinState struct {
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Vol        int64
	LastDayVol int64
	Mux        sync.RWMutex
}

// SymLiveData contains meta data for crawled data
type SymLiveData struct {
	Sym          string
	LastMinState LastMinState
	OHLCVD       OHLCVD
	Mux          sync.RWMutex
}

// Counter for counting listeners
type Counter struct {
	Count int
	Mux   sync.RWMutex
}

var liveFlags = flag.String("addr", "streamer.finance.yahoo.com:443", "http service address")

// SymLiveDataLookup for retrieving live data for processing by algos
var SymLiveDataLookup = map[string]*SymLiveData{}

// SubscribeLive listens to websocket for data from a list of indices and sends the accumulated data to a list of queues every minute
func SubscribeLive(indices []string, algoProcessingQueues []chan<- *SymLiveData) {
	ind2 := []string{}
	for i := range indices {
		sym := indices[i]
		println(i, "Started", sym)
		ind2 = append(ind2, sym)
		lms := LastMinState{}
		symData := SymLiveData{Sym: sym, OHLCVD: OHLCVD{}, LastMinState: lms}
		SymLiveDataLookup[sym] = &symData
		go updateLiveSymData(&symData, algoProcessingQueues)
	}

	// Start reading socket data...
	subCounters := []*Counter{}
	println("Total subscription count:", len(ind2))
	for i := 0; i < len(ind2); i += socketMaxSubscriptionCount {
		counter := Counter{}
		mx := i + socketMaxSubscriptionCount
		if len(ind2) < mx {
			mx = len(ind2)
		}
		go updateLastMinStateFromWebsocket(ind2[i:mx], &counter, SymLiveDataLookup)
		subCounters = append(subCounters, &counter)
	}
	go printSubscriptionRate(subCounters, 10)
}

func printSubscriptionRate(counters []*Counter, intSeconds int) {
	for range time.Tick(time.Duration(intSeconds) * time.Second) {
		cnt := 0
		for c := range counters {
			counters[c].Mux.RLock()
			cnt += counters[c].Count
			counters[c].Count = 0
			counters[c].Mux.RUnlock()
		}
		cnt /= intSeconds
		log.Print("Subscription message rate(/s): ", strconv.Itoa(cnt))
	}
}

// updateLastMinStateFromWebsocket Each socket can subscribe up to 50 before degrading
func updateLastMinStateFromWebsocket(symArr []string, counter *Counter, symDataLookup map[string]*SymLiveData) {
	u := url.URL{Scheme: "wss", Host: *liveFlags}
	//log.Printf("connecting to %s", u.String())

	header := http.Header{}

	header.Add("Host", "streamer.finance.yahoo.com")
	header.Add("Pragma", "no-cache")
	header.Add("Cache-Control", "no-cache")
	header.Add("Origin", "https://finance.yahoo.com")
	header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.75 Safari/537.36")
	header.Add("Accept-Encoding", "gzip, deflate, br")
	header.Add("Accept-Language", "en-GB,en;q=0.9,zh-CN;q=0.8,zh;q=0.7,en-US;q=0.6")

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	syms := strings.Join(symArr, "\",\"")
	body := []byte("{\"subscribe\":[\"" + syms + "\"]}")
	c.WriteMessage(websocket.TextMessage, body)

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println(time.Now().Unix(), "Reopening websocket...", err.Error())
			c.Close()
			updateLastMinStateFromWebsocket(symArr, counter, symDataLookup)
		}
		//log.Printf("recv: %s", message)
		base64.StdEncoding.Decode(message, message)
		d := parseProtobuf(message)

		t1 := time.Unix(0, d.Time*1E6)
		t2 := t1.Add(-12 * time.Hour)
		t2h := t2.Hour()
		t2m := t2.Minute()

		sym := d.Id

		if t2h > 9 && t2h < 16 || t2h == 9 && t2m >= 30 {
			px := float64(d.Price)
			symDataLookup[sym].LastMinState.Mux.Lock()
			if px > symDataLookup[sym].LastMinState.High {
				symDataLookup[sym].LastMinState.High = px
			}
			if px < symDataLookup[sym].LastMinState.Low || symDataLookup[sym].LastMinState.Open == 0 {
				symDataLookup[sym].LastMinState.Low = px
			}
			if symDataLookup[sym].LastMinState.Open == 0 {
				symDataLookup[sym].LastMinState.Open = px
			}
			symDataLookup[sym].LastMinState.Close = px
			symDataLookup[sym].LastMinState.Vol += d.DayVolume - symDataLookup[sym].LastMinState.LastDayVol //Negative vol!
			if symDataLookup[sym].LastMinState.Vol < 0 {
				symDataLookup[sym].LastMinState.Vol = 0
			}
			symDataLookup[sym].LastMinState.LastDayVol = d.DayVolume
			if debug {
				println(sym, "Price:", fmt.Sprintf("%f", px), "Unix time:", t1.Unix())
			}
			(*counter).Mux.Lock()
			(*counter).Count++
			(*counter).Mux.Unlock()

			symDataLookup[sym].LastMinState.Mux.Unlock()
		}
	}
}

func updateLiveSymData(symData *SymLiveData, algoProcessingQueues []chan<- *SymLiveData) {
	// This part attempts to append minute interval data object with latest retrieved data every minute
	for range time.Tick(time.Duration(1) * time.Minute) {
		time1 := time.Now()
		secs := time1.Second()
		time2 := time1.Add(-time.Duration(secs) * time.Second)

		if (*symData).LastMinState.Open > 0 {
			(*symData).Mux.Lock()
			lms := (*symData).LastMinState

			(*symData).OHLCVD.Ohlc = append([][]float64{[]float64{lms.Open, lms.High, lms.Low, lms.Close}}, (*symData).OHLCVD.Ohlc...)
			(*symData).OHLCVD.Date = append([]int64{time2.Unix()}, (*symData).OHLCVD.Date...)
			(*symData).OHLCVD.Vol = append([]int64{lms.Vol}, (*symData).OHLCVD.Vol...)
			(*symData).OHLCVD.LastTime = time2.Unix() // time represents time start of a min data
			(*symData).OHLCVD.Length++

			(*symData).LastMinState.Open = 0
			(*symData).LastMinState.High = 0
			(*symData).LastMinState.Low = 0
			(*symData).LastMinState.Close = 0
			(*symData).LastMinState.Vol = 0

			(*symData).Mux.Unlock()
			for ii := range algoProcessingQueues {
				algoProcessingQueues[ii] <- symData
			}
		}
	}
}

func parseProtobuf(dbody []uint8) yfinLive.YfinWS { //Warning, this can only work on indices as there are extra fields for other types not accounted here
	// You need to decode first, base64.StdEncoding.Decode(message, message)
	yfinWS := &yfinLive.YfinWS{}

	if err := proto.Unmarshal(dbody, yfinWS); err != nil { // i think it is safe to ignore errs
		return *yfinWS
	}
	return *yfinWS
}

func getUSTime() time.Time {
	return time.Now().Add(-12 * time.Hour)
}

func getMinutesSinceLastUSOpenTime() int {
	x := getUSTime()
	weekday := x.Weekday()
	if weekday > 5 {
		println("Today is weekend")
		return -1
	}
	hour := x.Hour()
	if hour == 9 && x.Minute() >= 30 {
		return x.Minute() - 30
	} else if hour < 16 && hour > 9 {
		return (hour-9)*60 + x.Minute() - 30
	}
	return -1
}
