# Go API Server for Price

Price 관련 정보를 조회하는 엔진입니다.

## Overview
- API version: v1
- server scheduler: 30s - 최신 가격정보 업데이트

### Running the server
서버 실행을 위해 다음 과정이 필요합니다:

#### 1) git repo 복제
```
git clone https://github.com/ypMarkJo/wmLight.git   
```

#### 2) DB 연결 설정
- wmLight/db/controller/connection.go 파일에 DB 정보 관련 const변수 수정

#### 3) DDL 실행
```
CREATE TABLE `latest_price` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `symbol` char(255) NOT NULL,
  `price` double NOT NULL,
  `source` char(255) NOT NULL,
  `timestamp` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

#### 4) 서버 실행

```
go mod tidy
```

```
go run main.go
```

### Test API
#### 1) 토큰 이름으로 최신의 토큰 정보 조회
요청:
  > localhost:8080/getprice/ETHUSD
  
응답: 
  >  [
      {
          "id": 205,
          "symbol": "ETHUSD",
          "price": 1862.3,
          "source": "bitfinex",
          "timestamp": "2023-07-31T19:16:02Z"
      }
  ]

#### 2) 토큰이름+가격출처로최신의토큰정보조회
요청:
  > localhost:8080/getprice/{symbol}/source/{source}
  
응답: 
  > [
    {
        "id": 222,
        "symbol": "USTUSD",
        "price": 1.0014,
        "source": "bitfinex",
        "timestamp": "2023-07-31T19:19:36Z"
    }
]

#### 3) 특정시간구간이주어졌을때해당시간동안의평균가격조회
요청:
  > localhost:8080/getavgprice/{symbol}?startTime=2023-07-31T18:13:30Z&endTime=2023-07-31T18:30:30Z
  
응답: 
  > {
    "average_value": 1862.3833333333332
}
