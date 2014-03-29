// Copyright (c) 2013-2014 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package donna

import()

// Search principal variation.
func (p *Position) searchPrincipal(alpha, beta, depth int) int {
        p.game.nodes++
        if depth == 0 {
                return p.searchQuiescence(alpha, beta)
        }

        ply := Ply()
        if ply > MaxDepth {
                return p.Evaluate()
        }

        if p.isRepetition() {
                return 0
        }

        // Checkmate pruning.
        if Checkmate - ply <= alpha {
                return alpha
        } else if ply - Checkmate >= beta {
                return beta
        }

        gen := p.StartMoveGen(ply)
        if !p.isInCheck(p.color) {
                gen.GenerateMoves()
        } else {
                gen.GenerateEvasions()
        }
        gen.rank()

        moveCount := 0
        bestMove, bestScore := Move(0), ply - Checkmate
        for move := gen.NextMove(); move != 0; move = gen.NextMove() {
                if position := p.MakeMove(move); position != nil {
                        //Log("%*sprin/%s> depth: %d, ply: %d, move: %s\n", ply*2, ` `, C(p.color), depth, ply, move)
                        inCheck := position.isInCheck(position.color)
                        reducedDepth := depth - 1
                        if inCheck {
                                reducedDepth++
                        }

                        moveScore := 0
                        if moveCount == 0 { // First move: follow principal variation.
                                moveScore = -position.searchPrincipal(-beta, -alpha, reducedDepth)
                        } else {
                                if reducedDepth == 0 {
                                        moveScore = -position.searchQuiescence(-alpha - 1, -alpha)
                                } else if inCheck {
                                        moveScore = -position.searchInCheck(-alpha, reducedDepth)
                                } else {
                                        moveScore = -position.searchWithZeroWindow(-alpha, reducedDepth)
                                }
                                if moveScore > alpha {
                                        moveScore = -position.searchPrincipal(-beta, -alpha, reducedDepth)
                                }
                        }

                        moveCount++
                        position.TakeBack(move)

                        if moveScore > bestScore {
                                position.saveBest(ply, move)
                                if moveScore > alpha {
                                        if moveScore >= beta {
                                                if move & isCapture == 0 && move & isPromo == 0 && move != p.game.killers[ply][0] {
                                                        p.game.killers[ply][1] = p.game.killers[ply][0]
                                                        p.game.killers[ply][0] = move
                                                	p.game.goodMoves[move.piece()][move.to()] += depth * depth;
                                                        //Log("==> depth: %d, node %d, killers %s/%s\n", depth, node, p.killers[0], p.killers[1])
                                                }
                                                return moveScore
                                        }
                                        alpha = moveScore
                                        bestMove = move
                                }
                                bestScore = moveScore
                        }
                }
        } // next move.

        if moveCount == 0 { // Checkmate
                if p.isInCheck(p.color) {
                        return bestScore
                } else { // Stalemate
                        return 0
                }
        } else if bestMove != Move(0) && bestMove & isCapture == 0 && bestMove & isPromo == 0 && bestMove != p.game.killers[ply][0] {
                p.game.killers[ply][1] = p.game.killers[ply][0]
                p.game.killers[ply][0] = bestMove
        	p.game.goodMoves[bestMove.piece()][bestMove.to()] += depth * depth;
                //Log("--> depth: %d, node %d, killers %s/%s\n", depth, node, p.game.killers[ply][0], p.game.killers[ply][1])
        }

        return bestScore
}
