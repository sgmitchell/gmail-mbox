package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

type progressLog struct {
	logRate   time.Duration
	expTotal  int64
	completed *atomic.Int64
	lastTs    time.Time
	lastCount *atomic.Int64
}

func NewProgressLog(tickRate time.Duration, total int) *progressLog {
	if tickRate < 100*time.Millisecond {
		tickRate = time.Second
	}
	return &progressLog{
		logRate:   tickRate,
		expTotal:  int64(total),
		completed: &atomic.Int64{},
		lastTs:    time.Now(),
		lastCount: &atomic.Int64{},
	}
}

func (p *progressLog) Add() {
	p.completed.Add(1)
}

func (p *progressLog) Go(ctx context.Context) {
	t := time.NewTicker(p.logRate)
	doLog := func() {
		now := time.Now()
		cur := p.completed.Load()
		doneSinceLast := cur - p.lastCount.Swap(cur)
		dur := now.Sub(p.lastTs)
		p.lastTs = now
		totStr, pctStr := "???", "??.??%"
		if p.expTotal > 0 {
			totStr = fmt.Sprintf("%d", p.expTotal)
			pctStr = fmt.Sprintf("%2.2f%%", float64(100*cur)/float64(p.expTotal))
		}

		overallMsg := fmt.Sprintf("%7d / %7s (%s)", cur, totStr, pctStr)
		sinceLastMsg := fmt.Sprintf("%7d at %.0f/s", doneSinceLast, float64(doneSinceLast)/dur.Seconds())

		msg := fmt.Sprintf("%s - %s", overallMsg, sinceLastMsg)
		slog.Info(msg)
	}

	for {
		select {
		case <-ctx.Done():
			doLog()
			slog.Info("Done")
			return
		case <-t.C:
			doLog()
		}
	}
}
