package stocktracker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type StockApi struct {
	alphaVantageApiKey string
	httpClient         http.Client
}

type StockValue struct {
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

type StockResponse struct {
	Symbol     string
	MetaData   map[string]string
	TimeSeries map[time.Time]StockValue
}

type rawResponse struct {
	MetaData   map[string]string            `json:"Meta Data"`
	TimeSeries map[string]map[string]string `json:"Time Series (60min)"`
}

func NewStockApi(alphaVantageApiKey string) StockApi {
	httpClient := http.Client{
		Timeout: 15 * time.Second,
	}

	return StockApi{
		alphaVantageApiKey: alphaVantageApiKey,
		httpClient:         httpClient,
	}
}

func (s StockApi) Get(symbol string) (StockResponse, error) {
	resp, err := s.httpClient.Get(fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_INTRADAY&symbol=%s&interval=60min&apikey=%s", symbol, s.alphaVantageApiKey))
	if err != nil {
		return StockResponse{}, errors.New(fmt.Sprintf("Unable to get stock information for %s: %v", symbol, err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return StockResponse{}, errors.New(fmt.Sprintf("Unable to read response body: %v", err))
	}

	r := rawResponse{}
	if err := json.Unmarshal(body, &r); err != nil {
		return StockResponse{}, errors.New(fmt.Sprintf("Unable to process json response: %v", err))
	}

	sr := StockResponse{}
	sr.Symbol = symbol
	sr.MetaData = r.MetaData
	sr.TimeSeries = map[time.Time]StockValue{}

	for timeString, d := range r.TimeSeries {
		openValue, err := strconv.ParseFloat(d["1. open"], 64)
		if err != nil {
			continue
		}

		highValue, err := strconv.ParseFloat(d["2. high"], 64)
		if err != nil {
			continue
		}

		lowValue, err := strconv.ParseFloat(d["3. low"], 64)
		if err != nil {
			continue
		}

		closeValue, err := strconv.ParseFloat(d["4. close"], 64)
		if err != nil {
			continue
		}

		volumeValue, err := strconv.ParseFloat(d["5. volume"], 64)
		if err != nil {
			continue
		}

		sv := StockValue{
			Open:   openValue,
			High:   highValue,
			Low:    lowValue,
			Close:  closeValue,
			Volume: volumeValue,
		}

		// fmt.Println(timeString + "-05:00")
		stockTime, err := time.Parse("2006-01-02 15:04:05-0700", timeString+"-0500")
		if err != nil {
			continue
		}

		sr.TimeSeries[stockTime] = sv
	}

	return sr, nil
}
