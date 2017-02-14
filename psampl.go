// Package psampl implements a generator of samples given a discrete probability distribution function
//
// Random numbers are taken from a source, which can be a pseudo-number generator
// like the one implemented by math/rand or a secure crypto source like the one of crypto/rand
// Once a distribution is set up by NewDistrib, samplers can be created (BiasSource), each associated with
// a random number generator source. BiasSource is safe for concurrent use by multiple goroutines given
// that the source of random numbers is itself safe.
//
// The algorithm used is Vose's alias method
// https://web.archive.org/web/20131029203736/http://web.eecs.utk.edu/~vose/Publications/random.pdf
// which is O(1) in generation time and O(1) in its use of input random numbers, but which has a
// setup time and memory of O(N) where N is the number of possible values for the samples.

package psampl

import (
	"errors"
	"fmt"
	"github.com/wadey/cryptorand"
	"math"
	"math/rand"
)

const (
	nBuf = 64
)

// A Distrib represents a probability distribution and can be later used
// to create biased sources with that probability distribution
type Distrib struct {
	prob  []float64
	alias []int
}

const pSmall = 1e-10

type intFifo []int

func (s *intFifo) deQueue() int {
	e := (*s)[0]
	(*s) = (*s)[1:]
	return e
}

func (s *intFifo) queue(n int) {
	*s = append(*s, n)
}

type probTables struct {
	small intFifo
	large intFifo
}

func initPTables(prob []float64) (probT probTables, probRenorm []float64, err error) {
	pl := 0.0
	probT = probTables{
		small: make(intFifo, 0, len(prob)),
		large: make(intFifo, 0, len(prob)),
	}
	probRenorm = append(probRenorm, prob...)
	for i := range probRenorm {
		pl += prob[i]

		probRenorm[i] *= float64(len(probRenorm))
		if probRenorm[i] < 1.0 {
			probT.small.queue(i)
		} else {
			probT.large.queue(i)
		}
	}
	pl = 1.0 - pl
	if pl > pSmall {
		msg := fmt.Sprintf("probabilities do not add to 1: %f", pl)
		return probTables{}, nil, errors.New(msg) //probs shoul sum 1
	}
	return probT, probRenorm, nil
}

func (d *Distrib) saturateProbs(probIdx intFifo) {
	for len(probIdx) != 0 {
		g := probIdx.deQueue()
		d.prob[g] = 1.0
	}
}

// NewDistrib returns a Distrib given an array of probabilities
// representing the discrete probability distribution function (i.e. the histogram).
func NewDistrib(prob []float64) (d *Distrib, err error) {

	if prob == nil {
		panic("no probability for distribution sampler")
	}

	probT, pr, err := initPTables(prob)
	if err != nil {
		return nil, err
	}

	d = &Distrib{
		prob:  make([]float64, len(pr)),
		alias: make([]int, len(pr)),
	}

	for len(probT.small) != 0 && len(probT.large) != 0 {
		l := probT.small.deQueue()
		g := probT.large.deQueue()

		d.prob[l] = pr[l]
		d.alias[l] = g
		pr[g] = (pr[g] + pr[l]) - 1
		if pr[g] < 1 {
			probT.small.queue(g)
		} else {
			probT.large.queue(g)
		}
	}
	d.saturateProbs(probT.large)
	d.saturateProbs(probT.small) //num instability...
	return d, nil
}

type BiasSource struct {
	d           *Distrib
	nBytesSampl int
	rsrc        *rand.Rand
}

// NewBiasSource biases a random source (expected to be uniformly distributed) using Distrib
// and creates BiasSource which can be used to obtain samples.
func (d *Distrib) NewBiasSource(rsrc *rand.Rand) *BiasSource {
	nb := bytesNeeded(len(d.prob))
	bs := &BiasSource{
		d:           d,
		nBytesSampl: nb,
		rsrc:        rsrc,
	}

	return bs
}

func biasCoin(prOne float64, rsrc *rand.Rand) bool {
	x := rsrc.Float64()
	return x < prOne
}

// SampleInt returns one sample from BiasSource, encoded as an int.
func (bs *BiasSource) SampleInt() (num int) {
	nt := len(bs.d.prob)
	i := bs.rsrc.Intn(nt)

	if biasCoin(bs.d.prob[i], bs.rsrc) {
		num = i
	} else {
		num = bs.d.alias[i]
	}
	return num
}

// Read fills p with packed samples in big endian.
// Each sample is at least one byte (it may occupy more than
// one byte, depends on the number of possible values of each sample)
func (bs *BiasSource) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i += bs.nBytesSampl {
		num := uint(bs.SampleInt())
		for j := 0; j < bs.nBytesSampl; j++ {
			nb := byte(num)
			p[i+bs.nBytesSampl-j-1] = nb
			num >>= 8
		}
	}
	return len(p), nil
}

func bytesNeeded(maxval int) int {
	return int(math.Log2(float64(maxval)))/8 + 1
}

// NewCryptoSampl is a helper function which creates a BiasSource
// out of Distrib with a secure number generator as origin of the input samples.
func (d *Distrib) NewCryptoSampl() *BiasSource {
	rsrc := rand.New(cryptorand.Source)
	bs := d.NewBiasSource(rsrc)
	return bs
}

// NewPrngSampl is a helper function which creates a sampler
// by Distrib with the standard pseudo-random number generator as origin of the input samples.
func (d *Distrib) NewPrngSampl(seed int64) *BiasSource {
	rsrc := rand.New(rand.NewSource(seed))
	bs := d.NewBiasSource(rsrc)
	return bs
}

type BiasBitSource struct {
	prOne float64
	rsrc  *rand.Rand
}

// NewBiasBitSource creates a source from which to sample biased bits.
// with the probability of getting one being prOne.
func NewBiasBitSource(prOne float64, rsrc *rand.Rand) *BiasBitSource {
	bbs := &BiasBitSource{
		prOne: prOne,
		rsrc:  rsrc,
	}
	return bbs
}

// SampleBit returns one sample from BiasBitSource, encoded as an bool.
func (bbs *BiasBitSource) SampleBit() bool {
	return biasCoin(bbs.prOne, bbs.rsrc)
}

func boolTouint(b bool) uint {
	i := uint(0)
	if b {
		i = 1
	}
	return i
}

// Read fills p with packed samples of BiasBitSource.
func (bbs *BiasBitSource) Read(p []byte) (n int, err error) {
	for i := range p {
		for j := uint(0); j < 8; j++ {
			bit := biasCoin(bbs.prOne, bbs.rsrc)
			b := boolTouint(bit)
			p[i] |= byte(b << j)
		}
	}
	return len(p), nil
}
