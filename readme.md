# 1. Overview
This crawler listens to Yahoo Finance websocket to retrieve realtime data with a partially decoded protobuf structure.

# 1.1. Design
The design leverages on goroutines for parallelized operations so as to be able to handle realtime data for high number of indices(>1k) with low internal latency.

![alt text](https://raw.githubusercontent.com/PhantomV1989/yfin_live_crawler/master/design.png)

## (1) 
The livedatacrawler.SubscribeLive subscribes up to 50 indices per each Yahoo finance websocket opened. This is to prevent overloading each websocket connection which will result in connection problems.

The data retrieved from Yahoo Finance websocket is encoded using protobuf(https://developers.google.com/protocol-buffers). Hence, decoding is necessary. A partially decoded(just enough to accomodate each stock's tick data) can be found at yfinLive/yfinLive.pb.go.


## (2)
This step stores the latest tick data collected per index in an internal object OHLCVD every minute. This minute interval historical data will be used for further data processing.

## (3)
Each index historical data is sent to a list of channels for data processing. Each of this channel should have a listener(trading algorithm) to retrieve the latest minute interval data.

# 2. Setup
`Environment: go1.13.8 linux/amd64`
In project folder, open terminal to install dependencies
```sh
go get
```

# 3. Examples
`***NOTE! You can only see outputs if you are running the examples when the US market is open(9.30pm SGT onwards)!!`
Most examples are in the my_test.go

To see the websocket outputs, run the test **TestSubscriptionWithDebug**. You should see the following
```sh
0 Started AAPL
1 Started GE
2 Started TSLA
3 Started F
...
Total subscription count: 400
BDX Price: 231.789993 Unix time: 1600458617
PBCT Price: 10.620000 Unix time: 1600458617
MT Price: 13.865000 Unix time: 1600458617
JWN Price: 14.579500 Unix time: 1600458616
SABR Price: 6.895000 Unix time: 1600458617
DDOG Price: 85.750000 Unix time: 1600458617
TSM Price: 80.449997 Unix time: 1600458616
...
```



