package controller

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"github.com/ypMarkJo/wmLight/src/api/model"
	"github.com/ypMarkJo/wmLight/src/config"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	BSC_TESTNET_URL = "https://data-seed-prebsc-1-s1.binance.org:8545"
	DAI_CONTRACT    = "0x0630521aC362bc7A19a4eE44b57cE72Ea34AD01c"
	ETH_CONTRACT    = "0x143db3CEEfbdfe5631aDD3E50f7614B6ba708BA7"
)

func SetNewPricesCL() {
	client, err := ethclient.Dial(BSC_TESTNET_URL)
	if err != nil {
		log.Fatal(err)
	}
	// chainlink smartcontract로 가격정보 조회 후 저장
	// 스마트 컨트랙트 ABI 로드
	abiFile, err := os.Open("src/abi/ChainLink.json")
	if err != nil {
		log.Fatal(err)
	}
	defer abiFile.Close()

	scanner := bufio.NewScanner(abiFile)
	var contractABIString string
	for scanner.Scan() {
		contractABIString += scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	contractABI, err := abi.JSON(strings.NewReader(contractABIString))
	if err != nil {
		log.Fatal(err)
	}

	contractAddresses := []string{DAI_CONTRACT, ETH_CONTRACT}
	symbols := []string{"DAI", "ETH"}

	for i := 0; i < len(contractAddresses); i++ {
		contractAddress := common.HexToAddress(contractAddresses[i])

		// 	스마트 컨트랙트 함수 호출
		data, err := contractABI.Pack("latestRoundData")
		if err != nil {
			log.Fatal(err)
		}

		callMsg := ethereum.CallMsg{
			To:   &contractAddress,
			Data: data,
		}
		result, err := client.CallContract(context.Background(), callMsg, nil)
		if err != nil {
			log.Fatal(err)
		}

		var roundData model.RoundData
		err = contractABI.UnpackIntoInterface(&roundData, "latestRoundData", result)
		if err != nil {
			log.Fatal(err)
		}
		// 블록 정보 조회
		header, err := client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			log.Fatal(err)
		}
		timestamp := header.Time
		bigFloatPrice := new(big.Float).SetInt(roundData.Answer)
		// big.Float를 float64로 변환
		float64Price, _ := bigFloatPrice.Float64()

		priceModel := &model.Price{
			Symbol:    symbols[i],
			Price:     float64Price,
			Source:    "chainlink",
			TimeStamp: time.Unix(int64(timestamp), 0),
		}

		_, err = config.AppCtx.Db.DB.Exec("INSERT INTO latest_price (symbol, price, source, timestamp) VALUES (?, ?, ?, ?)", priceModel.Symbol, priceModel.Price, priceModel.Source, priceModel.TimeStamp)
		if err != nil {
			log.Fatal("데이터 삽입 실패:", err)
		}

		fmt.Printf("%s 최신 가격: %v wei\n", symbols[i], roundData.Answer)
		fmt.Printf("타임스탬프: %v\n", time.Unix(int64(timestamp), 0))

	}
}

func SetNewPricesBF() {
	// Bitfinex API 엔드포인트 설정
	URL := "https://api.bitfinex.com/v1/pubticker/"
	symbols := []string{"ETHUSD", "USTUSD"} // 조회할 토큰 심볼

	for i := 0; i < len(symbols); i++ {
		// API 요청 보내기
		response, err := http.Get(URL + symbols[i])
		if err != nil {
			log.Fatal("API 요청 실패:", err)
		}
		defer response.Body.Close()

		// 응답 데이터 읽기
		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal("응답 데이터 읽기 실패:", err)
		}

		// API 응답 데이터 파싱
		var ticker model.TickerResponse
		err = json.Unmarshal(body, &ticker)
		if err != nil {
			log.Fatal("응답 데이터 파싱 실패:", err)
		}
		// 문자열 타임스탬프를 float64로 변환
		timestamp, err := strconv.ParseFloat(ticker.Timestamp, 64)
		if err != nil {
			log.Fatal("타임스탬프 변환 실패:", err)
		}
		// 변환된 float64 타임스탬프 값을 time.Time으로 변환
		time := time.Unix(int64(timestamp), 0)
		floatPrice, err := strconv.ParseFloat(ticker.LastPrice, 64)
		if err != nil {
			fmt.Println("변환 실패:", err)
			return
		}
		priceModel := &model.Price{
			Symbol:    symbols[i],
			Price:     floatPrice,
			Source:    "bitfinex",
			TimeStamp: time,
		}

		_, err = config.AppCtx.Db.DB.Exec("INSERT INTO latest_price (symbol, price, source, timestamp) VALUES (?, ?, ?, ?)", priceModel.Symbol, priceModel.Price, priceModel.Source, priceModel.TimeStamp)
		if err != nil {
			log.Fatal("데이터 삽입 실패:", err)
		}

		// 가격 정보 출력
		fmt.Printf("%s의 최신 가격: %f\n타임스탬프: %v\n", symbols[i], floatPrice, time)
	}

}

func getLatestPriceInfoBySingle(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("/getprice/{symbol} GET started!\n")
	// 파라미터 정보
	pathParameters := mux.Vars(r)

	// symbol nil 여부 확인
	symbol := pathParameters["symbol"]
	if symbol == "" {
		fmt.Printf("Failed to read path parameter")
		http.Error(w, "Missing 'symbol' parameter", http.StatusBadRequest)
		return
	}

	rows, err := config.AppCtx.Db.DB.Query(`SELECT id,symbol,price,source,timestamp FROM latest_price WHERE symbol=? ORDER BY timestamp DESC LIMIT 1`, symbol)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "쿼리 실행 실패", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var prices []model.Price
	for rows.Next() {
		var price model.Price
		var timestampStr string
		err := rows.Scan(&price.Id, &price.Symbol, &price.Price, &price.Source, &timestampStr)
		if err != nil {
			http.Error(w, "데이터 처리 실패", http.StatusInternalServerError)
			log.Fatal(err)
			return
		}
		price.TimeStamp, err = time.Parse("2006-01-02 15:04:05", timestampStr)
		if err != nil {
			http.Error(w, "시간 변환 실패", http.StatusInternalServerError)
			return
		}
		prices = append(prices, price)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prices)
}
func getLatestPriceInfoByDouble(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("/getprice/{symbol}/source/{source} GET started!\n")
	// 파라미터 정보
	pathParameters := mux.Vars(r)

	// symbol nil 여부 확인
	symbol := pathParameters["symbol"]
	if symbol == "" {
		fmt.Printf("Failed to read path parameter")
		http.Error(w, "Missing 'symbol' parameter", http.StatusBadRequest)
		return
	}
	// source nil 여부 확인
	source := pathParameters["source"]
	if source == "" {
		fmt.Printf("Failed to read path parameter")
		http.Error(w, "Missing 'source' parameter", http.StatusBadRequest)
		return
	}
	rows, err := config.AppCtx.Db.DB.Query(`SELECT id,symbol,price,source,timestamp FROM latest_price WHERE symbol=? and source=? ORDER BY timestamp DESC LIMIT 1`, symbol, source)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "쿼리 실행 실패", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var prices []model.Price
	for rows.Next() {
		var price model.Price
		var timestampStr string
		err := rows.Scan(&price.Id, &price.Symbol, &price.Price, &price.Source, &timestampStr)
		if err != nil {
			http.Error(w, "데이터 처리 실패", http.StatusInternalServerError)
			log.Fatal(err)
			return
		}
		price.TimeStamp, err = time.Parse("2006-01-02 15:04:05", timestampStr)
		if err != nil {
			http.Error(w, "시간 변환 실패", http.StatusInternalServerError)
			return
		}
		prices = append(prices, price)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prices)
}
func getPriceAverage(w http.ResponseWriter, r *http.Request) {

	// 시작시간과 종료시간을 HTTP 요청 매개변수로 받음
	startTime := r.URL.Query().Get("startTime")
	endTime := r.URL.Query().Get("endTime")
	fmt.Printf("/getavgprice/{symbol} GET started!\n")
	// 파라미터 정보
	pathParameters := mux.Vars(r)

	// symbol nil 여부 확인
	symbol := pathParameters["symbol"]
	if symbol == "" {
		fmt.Printf("Failed to read path parameter")
		http.Error(w, "Missing 'symbol' parameter", http.StatusBadRequest)
		return
	}

	rows, err := config.AppCtx.Db.DB.Query("SELECT AVG(price) FROM latest_price WHERE symbol=? AND timestamp BETWEEN ? AND ?", symbol, startTime, endTime)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "쿼리 실행 실패", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var averageValue float64

	for rows.Next() {
		err := rows.Scan(&averageValue)
		if err != nil {
			http.Error(w, "데이터 처리 실패", http.StatusInternalServerError)
			log.Fatal(err)
			return
		}
	}

	result := model.Result{AvgPrice: averageValue}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
