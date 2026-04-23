package null

import (
	"github.com/philpearl/plenc"
	"github.com/philpearl/plenc/plencnull"
)

//go:fix inline
func RegisterCodecs() {
	plencnull.RegisterCodecs()
}

//go:fix inline
func AddCodecs(p *plenc.Plenc) {
	plencnull.AddCodecs(p)
}
