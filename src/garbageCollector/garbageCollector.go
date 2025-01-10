package garbageCollector

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

func VramGC(vramGCCh chan [3]uint32, Cunt_ch *int) {
	for len(vramGCCh) > 0 {
		x := <-vramGCCh
		gl.DeleteVertexArrays(1, &x[0])
		gl.DeleteBuffers(1, &x[1])
		gl.DeleteBuffers(1, &x[2])
		//println(*Cunt_ch)
		*Cunt_ch--
	}
}
