package pinger

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Client はICMP通信を管理します
type Client struct {
	conn     net.PacketConn
	targetIP *net.IPAddr
	id       uint16
	seq      uint16
}

// NewClient はRaw Socketを開いて準備します
func NewClient(target string) (*Client, error) {
	// ip4:icmp を指定することで、IPヘッダはOSが処理し、ICMP部分だけ扱える
	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen (root権限が必要ですか?): %w", err)
	}

	dst, err := net.ResolveIPAddr("ip4", target)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// プロセスIDを識別子(ID)として利用
	return &Client{
		conn:     conn,
		targetIP: dst,
		id:       uint16(os.Getpid() & 0xffff),
		seq:      1,
	}, nil
}

// Close はソケットを閉じます
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// Ping は1回Pingを送信し、指定されたタイムアウト時間まで待ちます
// 返り値: (RTT, タイムアウトしたかどうか, エラー)
func (c *Client) Ping(timeout time.Duration) (time.Duration, bool, error) {
	currentSeq := c.seq
	c.seq++

	// --- 1. 送信準備 (Gopacket) ---
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true, // ICMPチェックサムを自動計算
	}

	// ICMPレイヤー作成
	icmpLayer := &layers.ICMPv4{
		TypeCode: layers.CreateICMPv4TypeCode(layers.ICMPv4TypeEchoRequest, 0),
		Id:       c.id,
		Seq:      currentSeq,
	}

	err := gopacket.SerializeLayers(buf, opts, icmpLayer,
		gopacket.Payload([]byte("RTO-PING")), // ペイロード
	)
	if err != nil {
		return 0, false, fmt.Errorf("serialize error: %w", err)
	}

	// --- 2. 送信 ---
	start := time.Now()
	_, err = c.conn.WriteTo(buf.Bytes(), c.targetIP)
	if err != nil {
		return 0, false, fmt.Errorf("send error: %w", err)
	}

	// --- 3. 受信 (タイムアウト付き) ---
	readBuf := make([]byte, 1500)
	deadline := start.Add(timeout)
	if err := c.conn.SetReadDeadline(deadline); err != nil {
		return 0, false, err
	}

	for {
		n, _, err := c.conn.ReadFrom(readBuf)
		if err != nil {
			// タイムアウト判定
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return 0, true, nil // タイムアウト発生
			}
			return 0, false, err // その他のエラー
		}

		// 受信時刻
		recvTime := time.Now()

		// --- 4. 解析 (Gopacket) ---
		// Raw Socket (ip4:icmp) で受信する場合、多くの環境でIPヘッダが除去されています。
		// まずICMPとして解析を試みます。
		packet := gopacket.NewPacket(readBuf[:n], layers.LayerTypeICMPv4, gopacket.Default)
		
		if icmpLayer := packet.Layer(layers.LayerTypeICMPv4); icmpLayer != nil {
			icmp, _ := icmpLayer.(*layers.ICMPv4)
			
			// EchoReply かつ IDとSeqが一致するか確認
			if icmp.TypeCode.Type() == layers.ICMPv4TypeEchoReply &&
				icmp.Id == c.id &&
				icmp.Seq == currentSeq {
				
				rtt := recvTime.Sub(start)
				return rtt, false, nil // 成功
			}
		}
		
		// 違うパケットだった場合はループして読み直し (期限切れまで)
	}
}

func (c *Client) CurrentSeq() uint16 {
	return c.seq
}