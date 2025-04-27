package transport

import (
	"net/http"
)

type WrapperFunc func(rt http.RoundTripper) http.RoundTripper

func Wrappers(fns ...WrapperFunc) WrapperFunc {
	if len(fns) == 0 {
		return nil
	}
	// optimize the common case of wrapping a possibly nil tr wrapper
	// with an additional wrapper
	if len(fns) == 2 && fns[0] == nil {
		return fns[1]
	}
	return func(rt http.RoundTripper) http.RoundTripper {
		base := rt
		for _, fn := range fns {
			if fn != nil {
				base = fn(base)
			}
		}
		return base
	}
}
