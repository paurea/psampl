# psampl
Package to generate biased random numbers.
Random numbers are taken from a source, which can be a pseudo-number generator
like the one implemented by math/rand or a secure crypto source like the one of crypto/rand
Once a distribution is set up by NewDistrib, samplers can be created (BiasSource), each associated with
a random number generator source. BiasSource is safe for concurrent use by multiple goroutines given
that the source of random numbers is itself safe.

The algorithm used is Vose's alias method
https://web.archive.org/web/20131029203736/http://web.eecs.utk.edu/~vose/Publications/random.pdf
which is O(1) in generation time and O(1) in its use of input random numbers, but which has a
setup time and memory of O(N) where N is the number of possible values for the samples.
