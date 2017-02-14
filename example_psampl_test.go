package psampl_test

import (
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"psampl"
)

const (
	nSamples      = 5
	nTotBytesRead = 16
	nBitsSample   = 10
	nValsSample   = 1 << (nBitsSample - 1)
	nBytesSample  = 2 //could be calculated from above
	nCoins        = 10
	prHead        = 0.1

	seedPrng = 7
)

var pdf = []float64{0.2, 0.6, 0.2}

func Example_Prng() {
	d, err := psampl.NewDistrib(pdf)
	if err != nil {
		log.Fatal("could not build Distrib %s:", err)
	}
	bs := d.NewPrngSampl(seedPrng)
	for i := 0; i < nSamples; i++ {
		n := bs.SampleInt()
		fmt.Println(n)
	}
}

func Example_Crypto() {
	d, err := psampl.NewDistrib(pdf)
	if err != nil {
		log.Fatal("could not build Distrib %s:", err)
	}
	bs := d.NewCryptoSampl()
	for i := 0; i < nSamples; i++ {
		n := bs.SampleInt()
		fmt.Println(n)
	}
}

func Example_Source() {
	rsrc := rand.New(rand.NewSource(seedPrng))
	d, err := psampl.NewDistrib(pdf)
	if err != nil {
		log.Fatal("could not build Distrib %s:", err)
	}
	bs := d.NewBiasSource(rsrc)
	for i := 0; i < nSamples; i++ {
		n := bs.SampleInt()
		fmt.Println(n)
	}
}

func Example_PrngRead() {
	bigPdf := make([]float64, nValsSample)
	for i := range bigPdf {
		bigPdf[i] = 1.0 / float64(nValsSample)
	}

	d, err := psampl.NewDistrib(bigPdf)
	if err != nil {
		log.Fatal("could not build Distrib %s:", err)
	}
	bs := d.NewPrngSampl(seedPrng)
	samples := make([]byte, nTotBytesRead)
	n, err := bs.Read(samples)
	//we internally guarantee the bytes
	if err != nil {
		log.Fatal("read failed")
	}
	// We have nValsSample possible values, may not fit in one byte.
	for i := 0; i < n; i += nBytesSample {
		n := uint(0)
		for j := 0; j < nBytesSample; j++ {
			n |= uint(samples[uint(i)+nBytesSample-uint(j)-1]) << uint(8*j)
		}
		fmt.Println(n)
	}
}

func Example_BitSample() {
	rsrc := rand.New(rand.NewSource(seedPrng))
	bbs := psampl.NewBiasBitSource(prHead, rsrc)
	nheads := 0
	for i := 0; i < nCoins; i++ {
		coin := bbs.SampleBit()
		if coin {
			nheads++
		}
	}
	fmt.Println(nheads)
}

func Example_BitRead() {
	rsrc := rand.New(rand.NewSource(seedPrng))
	bbs := psampl.NewBiasBitSource(prHead, rsrc)
	p := make([]byte, nCoins/8)
	n, err := bbs.Read(p)
	if err != nil {
		log.Fatal("read failed")
	}

	//Note: unpacking it as a big is fast, packing it is slow...
	bigN := big.NewInt(0)
	bigN = bigN.SetBytes(p)
	nheads := 0
	for i := 0; i < 8*n; i++ {
		coin := bigN.Bit(i) == 1
		if coin {
			nheads++
		}
	}
	fmt.Println(nheads)
}
