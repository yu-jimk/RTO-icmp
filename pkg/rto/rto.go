package rto

import (
	"math"
	"time"
)

const (
	// RFC 6298 推奨値
	minRTO = 1 * time.Second
	maxRTO = 60 * time.Second
	alpha  = 0.125 // 1/8
	beta   = 0.25  // 1/4
)

// Manager はRFC 6298に基づくRTOの状態を保持します
type Manager struct {
	SRTT   time.Duration
	RTTVAR time.Duration
	RTO    time.Duration
	first  bool
}

// NewManager は初期状態(RTO=1s)のマネージャーを作成します
func NewManager() *Manager {
	return &Manager{
		RTO:   minRTO,
		first: true,
	}
}

// Update はRTT計測成功時に呼ばれ、SRTT, RTTVAR, RTOを更新します
func (m *Manager) Update(rtt time.Duration) {
	if m.first {
		// (2.2) 最初の計測時
		m.SRTT = rtt
		m.RTTVAR = rtt / 2
		m.RTO = m.SRTT + 4*m.RTTVAR
		m.first = false
	} else {
		// (2.3) 2回目以降
		// RTTVAR = (1 - beta) * RTTVAR + beta * |SRTT - R'|
		diff := float64(m.SRTT - rtt)
		m.RTTVAR = time.Duration((1-beta)*float64(m.RTTVAR) + beta*math.Abs(diff))

		// SRTT = (1 - alpha) * SRTT + alpha * R'
		m.SRTT = time.Duration((1-alpha)*float64(m.SRTT) + alpha*float64(rtt))

		m.RTO = m.SRTT + 4*m.RTTVAR
	}

	m.clamp()
}

// Backoff はタイムアウト時に呼ばれ、RTOを2倍にします (Exponential Backoff)
func (m *Manager) Backoff() {
	m.RTO *= 2
	m.clamp()
}

// clamp はRTOを最小値・最大値の範囲に収めます
func (m *Manager) clamp() {
	if m.RTO < minRTO {
		m.RTO = minRTO
	}
	if m.RTO > maxRTO {
		m.RTO = maxRTO
	}
}