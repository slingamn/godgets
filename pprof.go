// Copyright (c) 2021 Shivaram Lingamneni
// released under the 0BSD license

package godgets

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

// be careful that the DefaultServeMux is not exposed anywhere else:
func StartPprofListener(address string) {
	go func() {
		log.Println(http.ListenAndServe(address, nil))
	}()
}
