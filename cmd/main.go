package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"rto-ping/pkg/pinger"
	"rto-ping/pkg/rto"
)

func main() {
	// 引数処理
	targetPtr := flag.String("t", "8.8.8.8", "Target IP address")
	flag.Parse()

	// 1. RTOマネージャーの初期化
	rtoMgr := rto.NewManager()

	// 2. Pingerクライアントの初期化 (Raw Socket)
	client, err := pinger.NewClient(*targetPtr)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fmt.Printf("Pinging %s with RFC 6298 RTO logic.\n", *targetPtr)
	fmt.Println("---------------------------------------------------")

	// メインループ
	for i := 0; i < 10; i++ {
		// 現在のRTOを取得
		currentTimeout := rtoMgr.RTO
		seq := client.CurrentSeq()

		fmt.Printf("[Seq %d] Sending... (Timeout limit: %v) ", seq, currentTimeout)

		// 送信と受信待機
		rtt, isTimeout, err := client.Ping(currentTimeout)
		
		if err != nil {
			// 通信エラー (ネットワーク不通など)
			fmt.Printf("\nError: %v\n", err)
		} else if isTimeout {
			// タイムアウト発生 -> RTO Backoff
			fmt.Printf("-> TIMEOUT!\n")
			rtoMgr.Backoff()
			fmt.Printf("   [RFC6298] Backing off. New RTO: %v\n", rtoMgr.RTO)
		} else {
			// 成功 -> RTO Update
			fmt.Printf("-> Reply! RTT: %v\n", rtt)
			rtoMgr.Update(rtt)
			fmt.Printf("   [RFC6298] Updated. SRTT: %v, RTTVAR: %v -> Next RTO: %v\n",
				rtoMgr.SRTT.Round(time.Microsecond),
				rtoMgr.RTTVAR.Round(time.Microsecond),
				rtoMgr.RTO.Round(time.Microsecond))
		}

		time.Sleep(1 * time.Second)
	}
}