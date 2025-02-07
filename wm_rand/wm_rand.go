package wm_rand

import (
	"errors"
	"math/big"
	"math/rand"
	"time"
)

type WmIdRand struct {
	l         int64
	r         int64
	cw        int64
	ccw       int64
	step      int64
	fr        int64
	t         int64
	s         int64
	certainty int
	random    *rand.Rand
	ms        int64
	d         bool
	ns        int64
}

func New(l, r int64) (*WmIdRand, error) {
	return NewWithCertainty(l, r, 200)
}

func NewWithCertainty(l, r int64, certainty int) (*WmIdRand, error) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	if l >= r {
		return nil, errors.New("origin must be non-negative and bound must be positive and origin must be less than bound")
	}
	s := r - l + 1
	var fr int64
	if big.NewInt(s).ProbablyPrime(certainty) {
		fr = r
	} else {
		fr = nextProbablePrime(big.NewInt(s), certainty).Int64() + l - 1
	}
	ns := fr - l
	step := random.Int63n(ns>>1) + 1
	for step != 1 && (ns+1)%step == 0 {
		fr = nextProbablePrime(big.NewInt(ns+1), certainty).Int64() + l - 1
		ns = fr - l
	}
	if fr <= l {
		return nil, errors.New("no prime number in the range")
	}
	cw := random.Int63n(ns) + l
	var ccw int64
	if cw-step < l {
		ccw = ns + cw - step + 1
	} else {
		ccw = cw - step
	}
	ms := rand.Int63n(r-l) + l
	d := random.Intn(2) == 0
	return &WmIdRand{
		l:         l,
		r:         r,
		cw:        cw,
		ccw:       ccw,
		step:      step,
		fr:        fr,
		t:         0,
		s:         s,
		certainty: certainty,
		random:    random,
		ms:        ms,
		d:         d,
		ns:        ns,
	}, nil
}

func (r *WmIdRand) Next() (int64, error) {
	if r.t >= r.s {
		return 0, errors.New("exhausted all possible values")
	}
	var v int64
	if r.random.Intn(2) == 0 {
		v = r.cw
		if r.cw > r.r {
			g := r.fr - r.cw
			r.cw = r.cw + (g - g%r.step)
			if r.cw+r.step > r.fr {
				r.cw = r.cw + r.step - r.ns - 1
			} else {
				r.cw = r.cw + r.step
			}
			v = r.cw
		}
		if r.cw+r.step > r.fr {
			r.cw = r.cw + r.step - r.ns - 1
		} else {
			r.cw = r.cw + r.step
		}
	} else {
		v = r.ccw
		if r.ccw > r.r {
			g := r.ccw - r.r - 1
			r.ccw = r.ccw - (g - g%r.step)
			if r.ccw-r.step < r.l {
				r.ccw = r.ns + r.ccw - r.step + 1
			} else {
				r.ccw = r.ccw - r.step
			}
			v = r.ccw
		}
		if r.ccw-r.step < r.l {
			r.ccw = r.ns + r.ccw - r.step + 1
		} else {
			r.ccw = r.ccw - r.step
		}
	}
	r.t++
	if r.d {
		if v <= r.ms {
			return v, nil
		}
		m := r.ms + ((r.r - r.ms) >> 1) + 1
		f := int64(0)
		if ((r.r - r.ms) & 1) == 0 {
			f = 1
		}
		return m - v - f + m, nil
	}
	if v > r.ms {
		return v, nil
	}
	f := int64(0)
	if ((r.ms - r.l + 1) & 1) == 0 {
		f = 1
	}
	m := r.l + ((r.ms - r.l) >> 1) + f
	return m - v - f + m, nil
}

func (r *WmIdRand) Reset() error {
	r.random = rand.New(rand.NewSource(time.Now().UnixNano()))
	r.step = r.random.Int63n(r.ns>>1) + 1
	if r.step != 1 && (r.ns+1)%r.step == 0 {
		r.fr = nextProbablePrime(big.NewInt(r.ns+1), r.certainty).Int64() + r.l - 1
		r.ns = r.fr - r.l
	}
	if r.fr <= r.l {
		return errors.New("fillBound illegal")
	}
	r.cw = r.random.Int63n(r.ns) + r.l

	if r.cw-r.step < r.l {
		r.ccw = r.ns + r.cw - r.step + 1
	} else {
		r.ccw = r.cw - r.step
	}
	r.t = 0
	r.ms = r.random.Int63n(r.r-r.l) + r.l
	r.d = r.random.Intn(2) == 0
	return nil
}

func nextProbablePrime(n *big.Int, certainty int) *big.Int {
	next := new(big.Int).Set(n)
	if next.Bit(0) == 0 {
		next.Add(next, big.NewInt(1))
	} else {
		next.Add(next, big.NewInt(2))
	}
	for {
		if next.ProbablyPrime(certainty) {
			return next
		}
		next.Add(next, big.NewInt(2))
	}
}
