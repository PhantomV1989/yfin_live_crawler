package livedatacrawler

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func processWorker(name string, processingQueue chan *SymLiveData) {
	for liveData := range processingQueue {
		if liveData.OHLCVD.LastTime != 0 {
			ohlc := liveData.OHLCVD.Ohlc[0]
			s := []string{
				liveData.Sym,
				liveData.OHLCVD.LastTimeString,
				fmt.Sprintf("%f", ohlc[0]),
				fmt.Sprintf("%f", ohlc[1]),
				fmt.Sprintf("%f", ohlc[2]),
				fmt.Sprintf("%f", ohlc[3]),
			}
			println(name + " processed: " + strings.Join(s, " "))
		}
	}
}

func TestWorker(t *testing.T) {
	var wg sync.WaitGroup
	var processingQueueA = make(chan *SymLiveData, 1000)
	wg.Add(1)
	go processWorker("worker_A", processingQueueA)
	processingQueueA <- &SymLiveData{Sym: "TSLA", OHLCVD: OHLCVD{LastTimeString: "2020-05-20 12:47:00",
		Ohlc: [][]float64{{1, 2, 3, 4}}}}
	wg.Wait()
}

func TestAAPLSubscription(t *testing.T) {
	// please test during market opening hours
	indices := []string{"AAPL"}
	var wg sync.WaitGroup
	_createQueue := func(s string) chan *SymLiveData {
		var processingQueueA = make(chan *SymLiveData, 100)
		go processWorker("worker_"+s, processingQueueA)
		return processingQueueA
	}
	myQueueList := []chan<- *SymLiveData{
		_createQueue("A"),
		_createQueue("B"),
		_createQueue("C"),
	}
	SubscribeLive(indices, myQueueList)
	wg.Add(1)
	wg.Wait()
}

func TestSubscriptionWithDebug(t *testing.T) {
	// please test during market opening hours
	indices := []string{"AAPL", "GE", "TSLA", "F", "NIO", "INO", "BAC", "AAL", "WFC", "AMD", "MSFT", "CCL", "DKNG", "VALE", "NCLH", "WKHS", "SRNE", "MRO", "ABEV", "CSCO", "PTON", "INTC", "SIRI", "PENN", "FB", "T", "UAL", "C", "KGC", "PLUG", "MU", "CLF", "CMCSA", "XOM", "FCX", "NOK", "OXY", "SAN", "BP", "DAL", "PBR", "PFE", "ITUB", "JPM", "SNAP", "KO", "BE", "BA", "COTY", "PCG", "VZ", "ORCL", "UBER", "GOLD", "BBD", "NVDA", "NKLA", "AUY", "M", "CTL", "MGM", "FISV", "ET", "TWTR", "WORK", "ERIC", "GM", "ZNGA", "VEON", "WUBA", "KHC", "ZM", "HBAN", "NKE", "MO", "EPD", "WMT", "CVX", "HRB", "MOS", "WMB", "KMI", "NLY", "LUV", "QCOM", "SLB", "BYND", "PDD", "RCL", "WRK", "CVS", "HL", "HPQ", "MDLZ", "MET", "AMCR", "BMY", "HAL", "MLHR", "PYPL", "PAA", "CHWY", "JD", "GILD", "COP", "ADT", "PBR-A", "MT", "MRNA", "BTG", "AMAT", "GGB", "RDS-A", "HPE", "ROKU", "SPCE", "KR", "JNJ", "MRK", "OPK", "CDE", "AES", "CNP", "BABA", "VER", "AMZN", "TECK", "DIS", "V", "EXC", "FSLY", "BBVA", "APA", "ATVI", "SNOW", "SBUX", "PINS", "RF", "PG", "BSX", "WBA", "KEY", "MS", "CX", "CFG", "NBL", "MPC", "SU", "TME", "NFLX", "JWN", "HMY", "BRK-B", "IMMU", "HBI", "EBAY", "APPS", "PM", "PEP", "SCHW", "PPL", "WPX", "TSM", "SQ", "DDOG", "GFI", "PE", "RKT", "RTX", "SABR", "LVS", "QRTEA", "PTAIF", "AEO", "TMUS", "NWSA", "TCEHY", "OSTK", "CLDR", "CMPGY", "DVN", "NEM", "LYG", "ABBV", "AA", "USB", "HST", "RDS-B", "BKR", "MAT", "CARR", "SDC", "BILI", "ALSMY", "JBLU", "VIAC", "KSS", "ABT", "BLDP", "GPS", "INFY", "CRM", "BAX", "LYFT", "GT", "RUN", "TXN", "HD", "NLOK", "HSBC", "SBSW", "MDT", "DOW", "VOD", "DOCU", "UAA", "PK", "NOV", "CSX", "GS", "OKE", "WY", "AIG", "ATUS", "FTCH", "TJX", "NVAX", "DD", "D", "FSLR", "CAT", "IVZ", "MTG", "ADBE", "APTV", "DBX", "ING", "TPR", "FTI", "FITB", "TEVA", "K", "WEN", "OVV", "AG", "CUK", "WDC", "AFL", "FDX", "MRVL", "TFC", "SO", "UPS", "EGO", "ETSY", "PPD", "SPR", "AXP", "IQ", "COG", "BHC", "AMGN", "AM", "LNC", "ADM", "W", "NUAN", "ADI", "AVTR", "FAST", "LOW", "MYL", "VLO", "SYF", "CTSH", "BCS", "EQT", "IBM", "FE", "DANOY", "HON", "NI", "IP", "AIV", "ANGI", "GLW", "KIM", "LI", "MPW", "MMM", "RRC", "SPG", "MCD", "TCOM", "STL", "LLY", "FERGY", "DB", "MAR", "AZN", "VST", "PSTG", "JCI", "NET", "CRWD", "VTR", "EOG", "AGNC", "DUK", "NUE", "NTNX", "CS", "ENPH", "CF", "CL", "CIG", "ETRN", "TDOC", "CZR", "DXC", "VIPS", "FHN", "UNM", "GRUB", "LEN", "PSX", "BK", "FOXA", "CNX", "LBTYK", "EQNR", "VICI", "UNH", "NVS", "TEF", "GSX", "CAH", "LB", "FIS", "ILMN", "GSK", "BMRN", "IBN", "MA", "CTVA", "SLM", "KKR", "SYY", "NG", "WU", "WM", "UNP", "GOOG", "OTIS", "XEL", "Z", "GPK", "UA", "WELL", "BLDR", "ALLY", "MNTA", "HWM", "EAF", "KDP", "FLEX", "ON", "PBCT", "GIS", "PEAK", "CPRI", "MPLX", "HOLX", "MXIM", "STLD", "COOP", "CVE", "O", "VFC", "TGTX", "ADP", "COST", "BDX", "MDLA", "XPEV", "IR", "SHOP", "AXTA", "NTAP", "AU", "HOG", "UMC", "PLD", "EXEL", "GPN", "ELAN", "TGT", "DISCA", "ARCC", "CGC"}
	var wg sync.WaitGroup
	_createQueue := func(s string) chan *SymLiveData {
		var processingQueueA = make(chan *SymLiveData, 100)
		go processWorker("worker_"+s, processingQueueA)
		return processingQueueA
	}
	myQueueList := []chan<- *SymLiveData{
		_createQueue("A"),
		_createQueue("B"),
		_createQueue("C"),
	}
	debug = true
	SubscribeLive(indices, myQueueList)
	wg.Add(1)
	wg.Wait()
}

func TestSubscriptionNoDebug(t *testing.T) {
	// please test during market opening hours
	indices := []string{"AAPL", "GE", "TSLA", "F", "NIO", "INO", "BAC", "AAL", "WFC", "AMD", "MSFT", "CCL", "DKNG", "VALE", "NCLH", "WKHS", "SRNE", "MRO", "ABEV", "CSCO", "PTON", "INTC", "SIRI", "PENN", "FB", "T", "UAL", "C", "KGC", "PLUG", "MU", "CLF", "CMCSA", "XOM", "FCX", "NOK", "OXY", "SAN", "BP", "DAL", "PBR", "PFE", "ITUB", "JPM", "SNAP", "KO", "BE", "BA", "COTY", "PCG", "VZ", "ORCL", "UBER", "GOLD", "BBD", "NVDA", "NKLA", "AUY", "M", "CTL", "MGM", "FISV", "ET", "TWTR", "WORK", "ERIC", "GM", "ZNGA", "VEON", "WUBA", "KHC", "ZM", "HBAN", "NKE", "MO", "EPD", "WMT", "CVX", "HRB", "MOS", "WMB", "KMI", "NLY", "LUV", "QCOM", "SLB", "BYND", "PDD", "RCL", "WRK", "CVS", "HL", "HPQ", "MDLZ", "MET", "AMCR", "BMY", "HAL", "MLHR", "PYPL", "PAA", "CHWY", "JD", "GILD", "COP", "ADT", "PBR-A", "MT", "MRNA", "BTG", "AMAT", "GGB", "RDS-A", "HPE", "ROKU", "SPCE", "KR", "JNJ", "MRK", "OPK", "CDE", "AES", "CNP", "BABA", "VER", "AMZN", "TECK", "DIS", "V", "EXC", "FSLY", "BBVA", "APA", "ATVI", "SNOW", "SBUX", "PINS", "RF", "PG", "BSX", "WBA", "KEY", "MS", "CX", "CFG", "NBL", "MPC", "SU", "TME", "NFLX", "JWN", "HMY", "BRK-B", "IMMU", "HBI", "EBAY", "APPS", "PM", "PEP", "SCHW", "PPL", "WPX", "TSM", "SQ", "DDOG", "GFI", "PE", "RKT", "RTX", "SABR", "LVS", "QRTEA", "PTAIF", "AEO", "TMUS", "NWSA", "TCEHY", "OSTK", "CLDR", "CMPGY", "DVN", "NEM", "LYG", "ABBV", "AA", "USB", "HST", "RDS-B", "BKR", "MAT", "CARR", "SDC", "BILI", "ALSMY", "JBLU", "VIAC", "KSS", "ABT", "BLDP", "GPS", "INFY", "CRM", "BAX", "LYFT", "GT", "RUN", "TXN", "HD", "NLOK", "HSBC", "SBSW", "MDT", "DOW", "VOD", "DOCU", "UAA", "PK", "NOV", "CSX", "GS", "OKE", "WY", "AIG", "ATUS", "FTCH", "TJX", "NVAX", "DD", "D", "FSLR", "CAT", "IVZ", "MTG", "ADBE", "APTV", "DBX", "ING", "TPR", "FTI", "FITB", "TEVA", "K", "WEN", "OVV", "AG", "CUK", "WDC", "AFL", "FDX", "MRVL", "TFC", "SO", "UPS", "EGO", "ETSY", "PPD", "SPR", "AXP", "IQ", "COG", "BHC", "AMGN", "AM", "LNC", "ADM", "W", "NUAN", "ADI", "AVTR", "FAST", "LOW", "MYL", "VLO", "SYF", "CTSH", "BCS", "EQT", "IBM", "FE", "DANOY", "HON", "NI", "IP", "AIV", "ANGI", "GLW", "KIM", "LI", "MPW", "MMM", "RRC", "SPG", "MCD", "TCOM", "STL", "LLY", "FERGY", "DB", "MAR", "AZN", "VST", "PSTG", "JCI", "NET", "CRWD", "VTR", "EOG", "AGNC", "DUK", "NUE", "NTNX", "CS", "ENPH", "CF", "CL", "CIG", "ETRN", "TDOC", "CZR", "DXC", "VIPS", "FHN", "UNM", "GRUB", "LEN", "PSX", "BK", "FOXA", "CNX", "LBTYK", "EQNR", "VICI", "UNH", "NVS", "TEF", "GSX", "CAH", "LB", "FIS", "ILMN", "GSK", "BMRN", "IBN", "MA", "CTVA", "SLM", "KKR", "SYY", "NG", "WU", "WM", "UNP", "GOOG", "OTIS", "XEL", "Z", "GPK", "UA", "WELL", "BLDR", "ALLY", "MNTA", "HWM", "EAF", "KDP", "FLEX", "ON", "PBCT", "GIS", "PEAK", "CPRI", "MPLX", "HOLX", "MXIM", "STLD", "COOP", "CVE", "O", "VFC", "TGTX", "ADP", "COST", "BDX", "MDLA", "XPEV", "IR", "SHOP", "AXTA", "NTAP", "AU", "HOG", "UMC", "PLD", "EXEL", "GPN", "ELAN", "TGT", "DISCA", "ARCC", "CGC"}
	var wg sync.WaitGroup
	_createQueue := func(s string) chan *SymLiveData {
		var processingQueueA = make(chan *SymLiveData, 100)
		go processWorker("worker_"+s, processingQueueA)
		return processingQueueA
	}
	myQueueList := []chan<- *SymLiveData{
		_createQueue("A"),
		_createQueue("B"),
		_createQueue("C"),
	}
	SubscribeLive(indices, myQueueList)
	wg.Add(1)
	wg.Wait()
}
