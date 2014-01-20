// Copyright (c) 2013 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package donna

import(`bytes`)

var tree [1024]Position
var node int

type Flags struct {
        enpassant     int       // En-passant square caused by previous move.
        irreversible  bool      // Is this position reversible?
}

type Position struct {
        game      *Game
        flags     Flags         // Flags set by last move leading to this position.
        pieces    [64]Piece     // Array of 64 squares with pieces on them.
        targets   [64]Bitmask   // Attack targets for pieces on each square of the board.
        board     [3]Bitmask    // [0] white pieces only, [1] black pieces, and [2] all pieces.
        attacks   [3]Bitmask    // [0] all squares attacked by white, [1] by black, [2] either white or black.
        outposts  [16]Bitmask   // Bitmasks of each piece on the board, ex. white pawns, black king, etc.
        count     [16]int       // counts of each piece on the board, ex. white pawns: 6, etc.
        color     int           // Side to make next move.
        stage     int           // Game stage (256 in the initial position).
        hash      uint64        // Polyglot hash value.
        inCheck   bool          // Is our king under attack?
        castles   uint8         // Castle rights mask.
}

func NewPosition(game *Game, pieces [64]Piece, color int, flags Flags) *Position {
        tree[node] = Position{ game: game, pieces: pieces, color: color }
        p := &tree[node]

        p.castles = castleKingside[White] | castleQueenside[White] |
                    castleKingside[Black] | castleQueenside[Black]

        if p.pieces[E1] != King(White) || p.pieces[H1] != Rook(White) {
                p.castles &= ^castleKingside[White]
        }
        if p.pieces[E1] != King(White) || p.pieces[A1] != Rook(White) {
                p.castles &= ^castleQueenside[White]
        }

        if p.pieces[E8] != King(Black) || p.pieces[H8] != Rook(Black) {
                p.castles &= ^castleKingside[Black]
        }
        if p.pieces[E8] != King(Black) || p.pieces[A8] != Rook(Black) {
                p.castles &= ^castleQueenside[Black]
        }

        return p.setupPieces().setupAttacks().computeStage()
}

func (p *Position) setupPieces() *Position {
        for square, piece := range p.pieces {
                if piece != 0 {
                        p.outposts[piece].set(square)
                        p.board[piece.color()].set(square)
                        p.count[piece]++
                }
        }
        p.board[2] = p.board[White] | p.board[Black]
        return p
}

func (p *Position) updatePieces(updates [64]Piece, squares []int) *Position {
        for _, square := range squares {
		if newer, older := updates[square], p.pieces[square]; newer != older {
			if older != 0 {
				p.count[older]--
			}
			p.outposts[older].clear(square)
			p.board[newer.color()^1].clear(square)
			p.pieces[square] = newer
	                if newer != 0 {
	                        p.outposts[newer].set(square)
	                        p.board[newer.color()].set(square)
				p.count[newer]++
	                } else {
	                        p.outposts[newer].clear(square)
	                        p.board[newer.color()].clear(square)
	                }
		}
        }
        p.board[2] = p.board[White] | p.board[Black]
        return p
}

func (p *Position) setupAttacks() *Position {
        board := p.board[2]
        for board != 0 {
                square := board.pop()
                piece := p.pieces[square]
                p.targets[square] = p.Targets(square)
                p.attacks[piece.color()].combine(p.targets[square])
        }
        //
        // Now that we have attack targets for both kings adjust them to make sure the
        // kings don't stomp on each other. Also, combine attacks bitmask and set the
        // flag is the king is being attacked.
        //
        p.updateKingTargets()
        p.attacks[2] = p.attacks[White] | p.attacks[Black]
        p.inCheck = p.isCheck(p.color)

        return p
}

func (p *Position) updateKingTargets() *Position {
	kingSquare := [2]int{ p.outposts[King(White)].first(), p.outposts[King(Black)].first() }

	if kingSquare[White] >= 0 && kingSquare[Black] >= 0 {
                p.targets[kingSquare[White]].exclude(p.targets[kingSquare[Black]])
                p.targets[kingSquare[Black]].exclude(p.targets[kingSquare[White]])
	}
        //
        // Add castle jump targets if castles are allowed.
        //
        if kingSquare[p.color] == homeKing[p.color] {
                if p.can00(p.color) {
                        p.targets[kingSquare[p.color]].set(kingSquare[p.color] + 2)
                }
                if p.can000(p.color) {
                        p.targets[kingSquare[p.color]].set(kingSquare[p.color] - 2)
                }
        }
        return p
}

func (p *Position) computeStage() *Position {
        p.hash  = p.polyglot()
        p.stage = 2 * (p.count[Pawn(White)]   + p.count[Pawn(Black)])   +
                  6 * (p.count[Knight(White)] + p.count[Knight(Black)]) +
                 12 * (p.count[Bishop(White)] + p.count[Bishop(Black)]) +
                 16 * (p.count[Rook(White)]   + p.count[Rook(Black)])   +
                 44 * (p.count[Queen(White)]  + p.count[Queen(Black)])
        return p
}

func (p *Position) MakeMove(move Move) *Position {
        eight := [2]int{ 8, -8 }
        color := move.color()
        flags := Flags{}

        from, to, piece, capture := move.split()
        delta := p.pieces
        delta[from] = 0
        delta[to] = piece
        squares := []int{ from, to }

        // Castle rights for current node are based on the castle rights from
        // the previous node.
        castles := tree[node].castles & castleRights[from] & castleRights[to]

        if kind := piece.kind(); kind == KING {
                if move.izCastle() {
                        flags.irreversible = true
                        switch to {
                        case G1:
                                delta[H1], delta[F1] = 0, Rook(White)
                                squares = append(squares, H1, F1)
                        case C1:
                                delta[A1], delta[D1] = 0, Rook(White)
                                squares = append(squares, A1, D1)
                        case G8:
                                delta[H8], delta[F8] = 0, Rook(Black)
                                squares = append(squares, H8, F8)
                        case C8:
                                delta[A8], delta[D8] = 0, Rook(Black)
                                squares = append(squares, A8, D8)
                        }
                }
                castles &= ^castleKingside[color]
                castles &= ^castleQueenside[color]
        } else {
                if kind == PAWN {
                        flags.irreversible = true
                        if move.izEnpassant() {//p.outposts[Pawn(color^1)]) {
                                //
                                // Mark the en-passant square.
                                //
                                flags.enpassant = from + eight[color]
                        } else if move.isEnpassantCapture(p.flags.enpassant) {
                                //
                                // Take out the en-passant pawn and decrement opponent's pawn count.
                                //
                                delta[to - eight[color]] = 0
				squares = append(squares, to - eight[color])
                        } else if promo := move.promo(); promo != 0 {
                                //
                                // Replace a pawn on 8th rank with the promoted piece.
                                //
                                delta[to] = promo
                        }
                }
                if p.castles & castleKingside[color] != 0 {
                        rookSquare := [2]int{ H1, H8 }
                        if delta[rookSquare[color]] != Rook(color) {
                                castles &= ^castleKingside[color]
                        }
                }
                if p.castles & castleQueenside[color] != 0 {
                        rookSquare := [2]int{ A1, A8 }
                        if delta[rookSquare[color]] != Rook(color) {
                                castles &= ^castleQueenside[color]
                        }
                }
        }
        if capture != 0 {
                flags.irreversible = true
                if p.castles & castleKingside[color^1] != 0 {
                        rookSquare := [2]int{ H1, H8 }
                        if delta[rookSquare[color^1]] != Rook(color^1) {
                                castles &= ^castleKingside[color^1]
                        }
                }
                if p.castles & castleQueenside[color^1] != 0 {
                        rookSquare := [2]int{ A1, A8 }
                        if delta[rookSquare[color]] != Rook(color^1) {
                                castles &= ^castleQueenside[color^1]
                        }
                }
        }

	node++
	tree[node] = Position{
		game:     p.game,
		board:    p.board,
		count:    p.count,
		pieces:   p.pieces,
		outposts: p.outposts,
		color:    color^1,
		flags:    flags,
		castles:  castles,
	}

        position := &tree[node]
	position.updatePieces(delta, squares).setupAttacks()
	if position.isCheck(color) {
                node--
		return nil
	}
	return position.computeStage()
}

func (p *Position) TakeBack(move Move) *Position {
        node--
        return &tree[node]
}

func (p *Position) isCheck(color int) bool {
        return p.outposts[King(color)] & p.attacks[color^1] != 0
}

func (p *Position) isRepetition() bool {
        if p.flags.irreversible {
                return false
        }

        for reps, prevNode := 1, node - 1; prevNode >= 0; prevNode-- {
                if tree[prevNode].flags.irreversible {
                        return false
                }
                if tree[prevNode].color == p.color && tree[prevNode].hash == p.hash {
                        reps++
                        if reps == 3 {
                                return true
                        }
                }
        }

        return false
}

func (p *Position) saveBest(ply int, move Move) {
        p.game.bestLine[ply][ply] = move
        p.game.bestLength[ply] = ply + 1
        for i := ply + 1; i < p.game.bestLength[ply + 1]; i++ {
                p.game.bestLine[ply][i] = p.game.bestLine[ply + 1][i]
                p.game.bestLength[ply]++
        }
}

func (p *Position) isPawnPromotion(piece Piece, target int) bool {
        return piece.isPawn() && ((piece.isWhite() && target >= A8) || (piece.isBlack() && target <= H1))
}

func (p *Position) can00(color int) bool {
        return p.castles & castleKingside[color] != 0 &&
               (gapKing[color] & p.board[2] == 0) &&
               (castleKing[color] & p.attacks[color^1] == 0)
}

func (p *Position) can000(color int) bool {
        return p.castles & castleQueenside[color] != 0 &&
               (gapQueen[color] & p.board[2] == 0) &&
               (castleQueen[color] & p.attacks[color^1] == 0)
}

// Compute position's polyglot hash.
func (p *Position) polyglot() (key uint64) {
        for i, piece := range p.pieces {
                if piece != 0 {
                        key ^= polyglotRandom[0:768][64 * piece.polyglot() + i]
                }
        }

	if p.castles & castleKingside[White] != 0 {
                key ^= polyglotRandom[768]
	}
	if p.castles & castleQueenside[White] != 0 {
                key ^= polyglotRandom[769]
	}
	if p.castles & castleKingside[Black] != 0 {
                key ^= polyglotRandom[770]
	}
	if p.castles & castleQueenside[Black] != 0 {
                key ^= polyglotRandom[771]
	}
        if p.flags.enpassant != 0 {
                col := Col(p.flags.enpassant)
                key ^= polyglotRandom[772 + col]
        }
	if p.color == White {
                key ^= polyglotRandom[780]
	}

	return
}

func (p *Position) String() string {
	buffer := bytes.NewBufferString("  a b c d e f g h")
        if !p.inCheck {
                buffer.WriteString("\n")
        } else {
                buffer.WriteString("  Check to " + C(p.color) + "\n")
        }
	for row := 7;  row >= 0;  row-- {
		buffer.WriteByte('1' + byte(row))
		for col := 0;  col <= 7; col++ {
			square := Square(row, col)
			buffer.WriteByte(' ')
			if piece := p.pieces[square]; piece != 0 {
				buffer.WriteString(piece.String())
			} else {
				buffer.WriteString("\u22C5")
			}
		}
		buffer.WriteByte('\n')
	}
	return buffer.String()
}
