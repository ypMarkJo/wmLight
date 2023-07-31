package main

import (
	"fmt"
	"github.com/ypMarkJo/wmLight/src/api/controller"
	"github.com/ypMarkJo/wmLight/src/config"
	dbController "github.com/ypMarkJo/wmLight/src/db/controller"
	"log"
	"net/http"
	"sync"
	"time"
)

func apiScheduler(wg *sync.WaitGroup) {
	// 30분마다 API를 실행하는 함수입니다.
	defer wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			controller.SetNewPricesCL()
			controller.SetNewPricesBF()
		}
	}
}

func main() {
	var wg sync.WaitGroup

	// 고루틴 실행 전 waitGroup에 1 추가
	wg.Add(1)

	// 	go apiScheduler를 고루틴으로 실행
	go apiScheduler(&wg)

	db, err := dbController.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()
	config.AppCtx.Db = db
	// 웹 서버 라우트 설정
	router := controller.NewRouter()

	// 웹 서버 시작
	fmt.Println("API 서버가 8080 포트에서 실행 중입니다.")
	http.ListenAndServe(":8080", router)

	wg.Wait()

}
