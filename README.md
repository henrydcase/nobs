# NOBS

Cryptographic implementation of Post-Quantum primitives in Go.

## Implemented primitives
* dh/
    - SIDH
* ec/
    - x448
* hash/
    - cSHAKE (sha3 coppied from "golang.org/x/crypto")
    - SM3
* rand/
    - CTR_DRBG with AES256 (NIST SP800-90A)
* kem/
    - SIKE: version 3 (as per paper on sike.org)
    
## Testing
```
make test
```

## Licence
WTFPL except if specified differently in subfolders
