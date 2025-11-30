package deduplicators

import (
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
)

// NewBloomFilter 创建基于布隆过滤器的 Deduplicator 实现
func NewBloomFilter(n uint, fp float64) *BloomFilter {
	return &BloomFilter{
		n:             n,
		fp:            fp,
		readonlyBloom: bloom.NewWithEstimates(n, fp),
		bloom:         bloom.NewWithEstimates(n, fp),
	}
}

// BloomFilter 基于布隆过滤器的 Deduplicator 实现
type BloomFilter struct {
	n  uint
	fp float64

	lock          sync.Mutex
	cnt           uint
	readonlyBloom *bloom.BloomFilter
	bloom         *bloom.BloomFilter
}

var _ Deduplicator = (*BloomFilter)(nil)

// Duplicate 校验是否重复的并记录下该内容
func (d *BloomFilter) Duplicate(data []byte) bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	ret := d.readonlyBloom.Test(data)
	if ret {
		return true
	}

	ret = d.readonlyBloom.TestOrAdd(data)
	if ret {
		return true
	}

	// 记录了一个新的数据
	d.cnt++

	// 检查更换过滤器
	if d.cnt >= d.n {
		d.readonlyBloom = d.bloom // 原过滤器转为只读，避免刚更换丢失历史数据
		d.bloom = bloom.NewWithEstimates(d.n, d.fp)
		d.cnt = 0
	}

	return false
}
