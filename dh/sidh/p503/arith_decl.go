// +build amd64,!noasm

package p503

import (
	. "github.com/henrydcase/nobs/dh/sidh/internal/isogeny"
)

// If choice = 0, leave x,y unchanged. If choice = 1, set x,y = y,x.
// If choice is neither 0 nor 1 then behaviour is undefined.
// This function executes in constant time.
//go:noescape
func fp503ConditionalSwap(x, y *FpElement, choice uint8)

// Compute z = x + y (mod p).
//go:noescape
func fp503AddReduced(z, x, y *FpElement)

// Compute z = x - y (mod p).
//go:noescape
func fp503SubReduced(z, x, y *FpElement)

// Compute z = x + y, without reducing mod p.
//go:noescape
func fp503AddLazy(z, x, y *FpElement)

// Compute z = x + y, without reducing mod p.
//go:noescape
func fp503X2AddLazy(z, x, y *FpElementX2)

// Compute z = x - y, without reducing mod p.
//go:noescape
func fp503X2SubLazy(z, x, y *FpElementX2)

// Compute z = x * y.
//go:noescape
func fp503Mul(z *FpElementX2, x, y *FpElement)

// Perform Montgomery reduction: set z = x R^{-1} (mod 2*p).
// Destroys the input value.
//go:noescape
func fp503MontgomeryReduce(z *FpElement, x *FpElementX2)

// Reduce a field element in [0, 2*p) to one in [0,p).
//go:noescape
func fp503StrongReduce(x *FpElement)
