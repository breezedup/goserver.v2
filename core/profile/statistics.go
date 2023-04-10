package profile

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/breezedup/goserver.v2/core/logger"
	"github.com/breezedup/goserver.v2/core/utils"
)

var TimeStatisticMgr = &timeStatisticMgr{
	elements: make(map[string]*timeElement),
}

type timeElement struct {
	name      string
	times     int64
	totalTick int64
	maxTick   int64
	minTick   int64
}

type timeStatisticMgr struct {
	elements map[string]*timeElement
	l        sync.RWMutex
}

func (this *timeStatisticMgr) WatchStart(name string) *TimeWatcher {
	tw := newTimeWatcher(name)
	return tw
}

func (this *timeStatisticMgr) addStatistic(name string, d int64) {
	this.l.RLock()
	if te, exist := this.elements[name]; exist {
		this.l.RUnlock()
		times := atomic.AddInt64(&te.times, 1)
		total := atomic.AddInt64(&te.totalTick, d)
		if d > te.maxTick {
			atomic.StoreInt64(&te.maxTick, d)
			if Config.SlowMS > 0 && d >= int64(Config.SlowMS)*int64(time.Millisecond) {
				logger.Logger.Warnf("###slow timespan name: %s  take:%s avg used:%s", strings.ToLower(te.name), utils.ToS(time.Duration(d)), utils.ToS(time.Duration(total/times)))
			}
		}
		if d < te.minTick {
			atomic.StoreInt64(&te.minTick, d)
		}

	} else {
		this.l.RUnlock()
		te := &timeElement{
			name:      name,
			times:     1,
			totalTick: d,
			maxTick:   d,
			minTick:   d,
		}
		this.l.Lock()
		this.elements[name] = te
		this.l.Unlock()
	}
}

func (this *timeStatisticMgr) dump(w io.Writer) {
	elements := make(map[string]*timeElement)
	this.l.RLock()
	for k, v := range this.elements {
		elements[k] = v
	}
	this.l.RUnlock()
	fmt.Fprintf(w, "| % -30s| % -10s | % -16s | % -16s | % -16s | % -16s |\n", "name", "times", "used", "max used", "min used", "avg used")
	for k, v := range elements {
		fmt.Fprintf(w, "| % -30s| % -10d | % -16s | % -16s | % -16s | % -16s |\n", strings.ToLower(k), v.times, utils.ToS(time.Duration(v.totalTick)), utils.ToS(time.Duration(v.maxTick)), utils.ToS(time.Duration(v.minTick)), utils.ToS(time.Duration(int64(v.totalTick)/v.times)))
	}
}
