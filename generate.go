// Copyright (c) 2013-2014 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package donna

import (
	`sort`
)

type MoveWithScore struct {
	move  Move
	score int
}

type MoveGen struct {
	p        *Position
	list     [128]MoveWithScore
	ply      int
	head     int
	tail     int
	pins     Bitmask
	obvious  Move
}

// Pre-allocate move generator array (one entry per ply) to avoid garbage
// collection overhead. Last entry serves for utility move generation, ex. when
// converting string notations or determining a stalemate.
var moveList [MaxPly+1]MoveGen

// Returns "new" move generator for the given ply. Since move generator array
// has been pre-allocated already we simply return a pointer to the existing
// array element re-initializing all its data.
func NewGen(p *Position, ply int) (gen *MoveGen) {
	gen = &moveList[ply]
	gen.p = p
	gen.list = [128]MoveWithScore{}
	gen.ply = ply
	gen.head, gen.tail = 0, 0
	gen.pins = p.pinnedMask(p.king[p.color])
	gen.obvious = Move(0)

	return gen
}

// Convenience method to return move generator for the current ply.
func NewMoveGen(p *Position) *MoveGen {
	return NewGen(p, ply())
}

// Returns new move generator for the initial step of iterative deepening
// (depth == 1) and existing one for subsequent iterations (depth > 1).
func NewRootGen(p *Position, depth int) *MoveGen {
	if depth == 1 {
		return NewGen(p, 0) // Zero ply.
	}
	return &moveList[0]
}

func (gen *MoveGen) reset() *MoveGen {
	gen.head = 0
	return gen
}

func (gen *MoveGen) size() int {
	return gen.tail
}

func (gen *MoveGen) onlyMove() bool {
	return gen.tail == 1
}

func (gen *MoveGen) NextMove() (move Move) {
	if gen.head < gen.tail {
		move = gen.list[gen.head].move
		gen.head++
	}
	return
}

// Returns true if the move is valid in current position i.e. it can be played
// without violating chess rules.
func (gen *MoveGen) isValid(move Move) bool {
	return gen.p.isValid(move, gen.pins)
}

// Removes invalid moves from the generated list. We use in iterative deepening
// to avoid filtering out invalid moves on each iteration.
func (gen *MoveGen) validOnly() *MoveGen {
	for move := gen.NextMove(); move != 0; move = gen.NextMove() {
		if !gen.isValid(move) {
			gen.remove()
		}
	}
	return gen.reset()
}

// Probes a list of generated moves and returns true if it contains at least
// one valid move.
func (gen *MoveGen) anyValid() bool {
	for move := gen.NextMove(); move != 0; move = gen.NextMove() {
		if gen.isValid(move) {
			return true
		}
	}
	return false
}

// Probes valid-only list of generated moves and returns true if the given move
// is one of them.
func (gen *MoveGen) amongValid(someMove Move) bool {
	for move := gen.NextMove(); move != 0; move = gen.NextMove() {
		if someMove == move {
			return true
		}
	}
	return false
}

// Assigns given score to the last move returned by the gen.NextMove().
func (gen *MoveGen) scoreMove(depth, score int) *MoveGen {
	current := &gen.list[gen.head - 1]

	//Log("-> Depth %d score %d current.score %d\n", depth, score, current.score)
	if depth == 1 || current.score == score + 1 {
		current.score = score
	} else if score != -depth || (score == -depth && current.score != score) {
		current.score += score // Fix up aspiration search drop.
	}
	//Log("=> Depth %d score %d current.score %d\n", depth, score, current.score)

	return gen
}

func (gen *MoveGen) rank(bestMove Move) *MoveGen {
	if gen.size() < 2 {
		return gen
	}

	for i := gen.head; i < gen.tail; i++ {
		move := gen.list[i].move
		if move == bestMove {
			gen.list[i].score = 0xFFFF
		} else if move & isCapture != 0 {
			gen.list[i].score = 8192 + move.value()
		} else if move == game.killers[gen.ply][0] {
			gen.list[i].score = 4096
		} else if move == game.killers[gen.ply][1] {
			gen.list[i].score = 2048
		} else {
			gen.list[i].score = game.good(move)
		}
	}

	sort.Sort(byScore{gen.list[gen.head:gen.tail]})
	return gen
}

func (gen *MoveGen) quickRank() *MoveGen {
	if gen.size() < 2 {
		return gen
	}

	for i := gen.head; i < gen.tail; i++ {
		if move := gen.list[i].move; move & isCapture != 0 {
			gen.list[i].score = 8192 + move.value()
		} else {
			gen.list[i].score = game.good(move)
		}
	}

	sort.Sort(byScore{gen.list[gen.head:gen.tail]})
	return gen
}

func (gen *MoveGen) rootRank(bestMove Move) *MoveGen {
	if gen.size() < 2 {
		return gen
	}

	// Find the best move and assign it the highest score.
	best, killer, semikiller, highest := -1, -1, -1, -Checkmate
	for i := gen.head; i < gen.tail; i++ {
		current := &gen.list[i]
		if current.move == bestMove {
			best = i
		} else if current.move == game.killers[gen.ply][0] {
			killer = i
		} else if current.move == game.killers[gen.ply][1] {
			semikiller = i
		}
		current.score += game.good(current.move) >> 3
		if current.score > highest {
			highest = current.score
		}
	}
	if best != -1 {
		gen.list[best].score = highest + 10
	}
	if killer != -1 {
		gen.list[killer].score = highest + 2
	}
	if semikiller != -1 {
		gen.list[semikiller].score = highest + 1
	}

	sort.Sort(byScore{gen.list[gen.head:gen.tail]})
	return gen
}

func (gen *MoveGen) add(move Move) *MoveGen {
	gen.list[gen.tail].move = move
	gen.tail++
	return gen
}

// Removes current move from the list by copying over the ramaining moves. Head and
// tail pointers get decremented so that calling NexMove() works as expected.
func (gen *MoveGen) remove() *MoveGen {
	copy(gen.list[gen.head-1:], gen.list[gen.head:])
	gen.head--
	gen.tail--
	return gen
}

// Returns an array of generated moves by continuously appending the NextMove()
// until the list is empty.
func (gen *MoveGen) allMoves() (moves []Move) {
	for move := gen.NextMove(); move != 0; move = gen.NextMove() {
		moves = append(moves, move)
	}
	gen.reset()

	return
}

// Sorting moves by their relative score based on piece/square for regular moves
// or least valuaeable attacker/most valueable victim for captures.
type byScore struct {
	list []MoveWithScore
}

func (her byScore) Len() int           { return len(her.list) }
func (her byScore) Swap(i, j int)      { her.list[i], her.list[j] = her.list[j], her.list[i] }
func (her byScore) Less(i, j int) bool { return her.list[i].score > her.list[j].score }
