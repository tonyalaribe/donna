// Copyright (c) 2013-2014 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package donna

import(`github.com/michaeldv/donna/expect`; `testing`)

func TestCache000(t *testing.T) {
	engine.cacheSize = 0.5
	p := NewGame().start()
	move := NewMove(p, E2, E4)
	p = p.makeMove(move).cache(move, 42, 1, cacheExact)

	cached := p.probeCache()
	expect.Eq(t, cached.move, move)
	expect.Eq(t, cached.score, 42)
	expect.Eq(t, cached.depth, 1)
	expect.Eq(t, cached.flags, uint8(cacheExact))
	expect.Eq(t, cached.hash, p.hash)
}
