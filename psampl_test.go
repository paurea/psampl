package psampl_test

import (
	"math"
	"math/rand"
	"psampl"
	"testing"
	"math/big"
)

const (
	nTestSampleInts = 100000
	nTestSampleBits = 1000000
	nTestSeeds      = 10
	nBenchSeeds     = 10
	errEps          = 0.01
	seedSec         = 0 //Special seed meaning use the secure pnrg in the tests

	nVals      = 357 //one byte and a half to see if it works
	nBytes     = 2
	nBytesRead = nTestSampleInts * nBytes

	prCoinStep = 0.1
)

func trotations(pr []float64, t *testing.T, seed int64) {
	for _ = range pr {
		//try for each rotation of pr
		d, err := psampl.NewDistrib(pr)
		if err != nil {
			t.Errorf("could not build Distrib %s:", err)
		}
		var bs *psampl.BiasSource
		if seed == seedSec {
			bs = d.NewCryptoSampl()
		} else {
			bs = d.NewPrngSampl(seed)
		}

		nf := make([]float64, len(pr))
		for i := 0; i < nTestSampleInts; i++ {
			n := bs.SampleInt()
			nf[n]++
		}
		for i, nf := range nf {
			p := float64(nf / nTestSampleInts)
			if math.Abs(p-pr[i]) > errEps {
				t.Errorf("obtained prob %f and should be %f, in prs: %v", p, pr[i], pr)
			}
		}
		pr = append(pr[1:], pr[0])
	}
}

func tread(t *testing.T, seed int64) {
	pr := make([]float64, nVals)
	for i := range pr {
		pr[i] = 1.0 / float64(nVals)
	}

	d, err := psampl.NewDistrib(pr)
	if err != nil {
		t.Errorf("could not build Distrib %s:", err)
	}
	var bs *psampl.BiasSource
	if seed == seedSec {
		bs = d.NewCryptoSampl()
	} else {
		bs = d.NewPrngSampl(seed)
	}

	nf := make([]float64, len(pr))
	samples := make([]byte, nBytesRead)
	n, err := bs.Read(samples)
	//we internally guarantee the bytes
	if err != nil || n != nBytesRead {
		t.Fatal("read failed")
	}
	for i := 0; i < len(samples); i += nBytes {
		n := uint(0)
		for j := 0; j < nBytes; j++ {
			n |= uint(samples[i+nBytes-j-1]) << uint(8*j)
		}
		nf[n]++
	}
	for i, nf := range nf {
		p := float64(nf / nTestSampleInts)
		if math.Abs(p-pr[i]) > errEps {
			t.Errorf("obtained prob %f and should be %f, in prs: %v", p, pr[i], pr)
		}
	}
}

//all sum 1
var prs = [][]float64{
	[]float64{0.09, 0.4, 0.01, 0.5},
	[]float64{0.5, 0.5},
	[]float64{1.0},
	[]float64{1e-9, 1 - 1e-9},
}

func TestSampleBit(t *testing.T) {
	seed := int64(1)
	rsrc := rand.New(rand.NewSource(seed))
	for pr := 0.0; pr < 1.0; pr += prCoinStep {
		bbs := psampl.NewBiasBitSource(pr, rsrc)
		nheads := 0
		for i := 0; i < nTestSampleBits; i++ {
			coin := bbs.SampleBit()
			if coin {
				nheads++
			}
		}
		spr := float64(nheads) / float64(nTestSampleBits)
		if math.Abs(pr-spr) > errEps {
			t.Errorf("bit not well biased %f and should be %f", spr, pr)
		}
	}
}

func TestBiasBitRead(t *testing.T) {
	seed := int64(1)
	rsrc := rand.New(rand.NewSource(seed))
	for pr := 0.0; pr < 1.0; pr += prCoinStep {
		bbs := psampl.NewBiasBitSource(pr, rsrc)
		p := make([]byte, nTestSampleBits/8)
		n, err := bbs.Read(p)
		//we internally guarantee the bytes
		if err != nil || n != nTestSampleBits/8 {
			t.Fatal("read failed")
		}
		//Note: unpacking it as a big is fast, packing it is slow...
		bigN := big.NewInt(0)
		bigN = bigN.SetBytes(p)
		nheads := 0
		for i := 0; i < nTestSampleBits; i++ {
			bit := bigN.Bit(i)
			if bit == 1 {
				nheads++
			}
		}
		spr := float64(nheads) / float64(nTestSampleBits)
		if math.Abs(pr-spr) > errEps {
			t.Errorf("bit not well biased %f and should be %f", spr, pr)
		}
	}
}

func TestSampleInt(t *testing.T) {
	//seed 0 is the crypto pnrg
	for seed := int64(0); seed < nTestSeeds; seed++ {
		for _, pr := range prs {
			trotations(pr, t, seed)
		}
	}
}

func TestBiasRead(t *testing.T) {
	//seed 0 is the crypto pnrg
	for seed := int64(0); seed < nTestSeeds; seed++ {
		tread(t, seed)
	}
}

func tbench(pr []float64, b *testing.B, seed int64) {
	b.StartTimer()
	//try for each rotation of pr
	d, err := psampl.NewDistrib(pr)
	if err != nil {
		b.Errorf("could not build Distrib %s:", err)
	}
	var bs *psampl.BiasSource
	if seed == seedSec {
		bs = d.NewCryptoSampl()
	} else {
		bs = d.NewPrngSampl(seed)
	}

	nf := make([]float64, len(pr))
	for i := 0; i < nTestSampleInts; i++ {
		n := bs.SampleInt()
		nf[n]++
	}
	b.StopTimer()
	for i, nf := range nf {
		p := float64(nf / nTestSampleInts)
		if math.Abs(p-pr[i]) > errEps {
			b.Errorf("obtained prob %f and should be %f, in prs: %v", p, pr[i], pr)
		}
	}
}

func BenchmarkCrypto(b *testing.B) {
	b.StopTimer()
	for _, pr := range prs {
		for _ = range pr {
			tbench(pr, b, seedSec)
			pr = append(pr[1:], pr[0])
		}
	}
}

func BenchmarkPnrg(b *testing.B) {
	b.StopTimer()
	for seed := int64(0); seed < nTestSeeds; seed++ {
		for _, pr := range prs {
			for _ = range pr {
				tbench(pr, b, seed)
				pr = append(pr[1:], pr[0])
			}
		}
	}
}
