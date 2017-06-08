# psampl go package
Package to generate biased random numbers.
Random numbers are taken from a source, which can be a pseudo-number generator
like the one implemented by math/rand or a secure crypto source like the one of crypto/rand.
Once a distribution is set up by NewDistrib, samplers can be created (BiasSource), each associated with
a random number generator source. BiasSource is safe for concurrent use by multiple goroutines given
that the source of random numbers used to create the Distrib is safe.
# Algorithm
The algorithm used is [Vose's alias method] (https://web.archive.org/web/20131029203736/http://web.eecs.utk.edu/~vose/Publications/random.pdf) which is O(1) in generation time and O(1) in its use of input random numbers, but which has a setup time and memory of O(N) where N is the number of possible values for the samples.

    go get github.com/paurea/psampl

```
import "github.com/paurea/psampl"
```
## Usage

#### type BiasBitSource

```
type BiasBitSource struct {
}
```


#### func  NewBiasBitSource

```
func NewBiasBitSource(prOne float64, rsrc *rand.Rand) *BiasBitSource
```
NewBiasBitSource creates a source from which to sample biased bits. with the
probability of getting one being prOne.

#### func (*BiasBitSource) Read

```
func (bbs *BiasBitSource) Read(p []byte) (n int, err error)
```
Read fills p with packed samples of BiasBitSource.

#### func (*BiasBitSource) SampleBit

```
func (bbs *BiasBitSource) SampleBit() bool
```
SampleBit returns one sample from BiasBitSource, encoded as an bool.

#### type BiasSource

```
type BiasSource struct {
}
```


#### func (*BiasSource) Read

```
func (bs *BiasSource) Read(p []byte) (n int, err error)
```
Read fills p with packed samples in big endian. Each sample is at least one byte
(it may occupy more than one byte, depends on the number of possible values of
each sample)

#### func (*BiasSource) SampleInt

```
func (bs *BiasSource) SampleInt() (num int)
```
SampleInt returns one sample from BiasSource, encoded as an int.

#### type Distrib

```
type Distrib struct {
}
```

A Distrib represents a probability distribution and can be later used to create
biased sources with that probability distribution

#### func  NewDistrib

```
func NewDistrib(prob []float64) (d *Distrib, err error)
```
NewDistrib returns a Distrib given an array of probabilities representing the
discrete probability distribution function (i.e. the histogram).

#### func (*Distrib) NewBiasSource

```
func (d *Distrib) NewBiasSource(rsrc *rand.Rand) *BiasSource
```
NewBiasSource biases a random source (expected to be uniformly distributed)
using Distrib and creates BiasSource which can be used to obtain samples.

#### func (*Distrib) NewCryptoSampl

```
func (d *Distrib) NewCryptoSampl() *BiasSource
```
NewCryptoSampl is a helper function which creates a BiasSource out of Distrib
with a secure number generator as origin of the input samples.

#### func (*Distrib) NewPrngSampl

```
func (d *Distrib) NewPrngSampl(seed int64) *BiasSource
```
NewPrngSampl is a helper function which creates a sampler by Distrib with the
standard pseudo-random number generator as origin of the input samples.
