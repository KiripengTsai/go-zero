package breaker

import (
	"math"
	"time"

	"github.com/tal-tech/go-zero/core/collection"
	"github.com/tal-tech/go-zero/core/mathx"
)

const (
	// 250ms for bucket duration
	window     = time.Second * 10
	buckets    = 40
	k          = 1.5
	protection = 5
)

// googleBreaker is a netflixBreaker pattern from google.
// see Client-Side Throttling section in https://landing.google.com/sre/sre-book/chapters/handling-overload/
type googleBreaker struct {
	k     float64                   // 倍值 默认 1.5
	stat  *collection.RollingWindow //滑动时间窗口，用来怼请求失败和成功计数
	proba *mathx.Proba              // 动态概率
}

func newGoogleBreaker() *googleBreaker {
	bucketDuration := time.Duration(int64(window) / int64(buckets))
	st := collection.NewRollingWindow(buckets, bucketDuration)
	return &googleBreaker{
		stat:  st,
		k:     k,
		proba: mathx.NewProba(),
	}
}

// max(0, requests - (K * accepts)/(requests + 1)) google Sre 过载保护算法
func (b *googleBreaker) accept() error {
	accepts, total := b.history()
	weightedAccepts := b.k * float64(accepts)
	// https://landing.google.com/sre/sre-book/chapters/handling-overload/#eq2101
	//计算丢弃请求概率
	dropRatio := math.Max(0, (float64(total-protection)-weightedAccepts)/float64(total+1))
	if dropRatio <= 0 {
		return nil
	}

	// 动态判断是否触发熔断
	if b.proba.TrueOnProba(dropRatio) {
		return ErrServiceUnavailable
	}

	return nil
}

func (b *googleBreaker) allow() (internalPromise, error) {
	if err := b.accept(); err != nil {
		return nil, err
	}

	return googlePromise{
		b: b,
	}, nil
}

// 调用doReq办法，在这个办法中首先通过accept 校验是否触发熔断
func (b *googleBreaker) doReq(req func() error, fallback func(err error) error, acceptable Acceptable) error {
	// 判断是否触发熔断
	if err := b.accept(); err != nil {
		if fallback != nil {
			return fallback(err)
		} else {
			return err
		}
	}

	defer func() {
		if e := recover(); e != nil {
			b.markFailure()
			panic(e)
		}
	}()

	// 执行调用函数
	err := req()
	//正常请求计算
	if acceptable(err) {
		b.markSuccess()
	} else {
		//异常请求计算
		b.markFailure()
	}

	return err
}

func (b *googleBreaker) markSuccess() {
	b.stat.Add(1)
}

func (b *googleBreaker) markFailure() {
	b.stat.Add(0)
}

func (b *googleBreaker) history() (accepts int64, total int64) {
	b.stat.Reduce(func(b *collection.Bucket) {
		accepts += int64(b.Sum)
		total += b.Count
	})

	return
}

type googlePromise struct {
	b *googleBreaker
}

func (p googlePromise) Accept() {
	p.b.markSuccess()
}

func (p googlePromise) Reject() {
	p.b.markFailure()
}
