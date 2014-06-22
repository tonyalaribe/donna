// Copyright (c) 2013-2014 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style license that can
// be found in the LICENSE file.

package donna

var (
	valuePawn      = Score{  100,  129 }
	valueKnight    = Score{  408,  423 } //  350,  330
	valueBishop    = Score{  418,  428 } //  355,  360
	valueRook      = Score{  635,  639 } //  525,  550
	valueQueen     = Score{ 1260, 1279 } // 1000, 1015

	rightToMove    = Score{   12,    5 }
	pawnBlocked    = Score{    2,    6 } //~~~
	bishopPair     = Score{   43,   56 } // Bonus for a pair of bishops.
	bishopPairPawn = Score{    4,    0 } // Penalty for each 5+ pawn when we've got a pair of bishops.
	bishopPawn     = Score{    4,    6 } // Penalty for each pawn on the same colored square as a bishop.
	bishopBoxed    = Score{   73,    0 } //~~~
	bishopDanger   = Score{   35,    0 } // Bonus when king is under attack and sides have opposite-colored bishops.
	rookOnPawn     = Score{    5,   14 }
	rookOnOpen     = Score{   22,   10 }
	rookOnSemiOpen = Score{    9,    5 }
	rookOn7th      = Score{    5,   10 }
	rookBoxed      = Score{   45,    0 }
	queenOnPawn    = Score{    2,   10 }
	queenOn7th     = Score{    1,    4 }
	behindPawn     = Score{    8,    0 }
	hangingAttack  = Score{   10,   12 }
	coverMissing   = Score{   45,    0 } //~~~ Missing cover pawn penalty.
	coverDistance  = Score{    8,    0 } //~~~ Cover pawn row distance from king penalty.
)

// Weight percentages applied to evaluation scores before computing the overall
// blended score.
var weights = []Score{
	{105, 134}, 	// [0] Mobility.
	{ 95,  79}, 	// [1] Pawn structure.
	{ 86, 107}, 	// [2] Passed pawns.
	{106,   0}, 	// [3] King safety.
	{120,   0}, 	// [4] Enemy's king safety.
}

// Piece values for calculating most valueable victim/least valueable attacker,
// indexed by piece.
var pieceValue = [14]int{
	0, 0,
	valuePawn.midgame,   valuePawn.midgame,
	valueKnight.midgame, valueKnight.midgame,
	valueBishop.midgame, valueBishop.midgame,
	valueRook.midgame,   valueRook.midgame,
	valueQueen.midgame,  valueQueen.midgame,
	0, 0,
}

// Piece/square table: gets initilized on startup from the bonus arrays below.
var pst = [14][64]Score{{},}

// Piece/square bonus points, visually arranged from White's point of view. The
// square index is used directly for Black and requires a flip for White.
var bonusPawn = [2][64]int{
	{  // vvvvvvvvvvvvvvvv Black vvvvvvvvvvvvvvvv
	        0,   0,   0,   0,   0,   0,   0,   0,
	      -10,  -3,   2,   7,   7,   2,  -3, -10,
	      -10,  -3,   4,   7,   7,   4,  -3, -10,
	      -10,  -3,   8,  17,  17,   8,  -3, -10,
	      -10,  -3,   8,  27,  27,   8,  -3, -10,
	      -10,  -3,   4,  17,  17,   4,  -3, -10,
	      -10,  -3,   2,   7,   7,   2,  -3, -10,
	        0,   0,   0,   0,   0,   0,   0,   0,
	}, {
	        0,   0,   0,   0,   0,   0,   0,   0,
	       -1,  -1,  -1,  -1,  -1,  -1,  -1,  -1,
	       -2,  -2,  -2,  -2,  -2,  -2,  -2,  -2,
	       -3,  -3,  -3,  -3,  -3,  -3,  -3,  -3,
	       -4,  -4,  -4,  -4,  -4,  -4,  -4,  -4,
	       -5,  -5,  -5,  -5,  -5,  -5,  -5,  -5,
	       -6,  -6,  -6,  -6,  -6,  -6,  -6,  -6,
	        0,   0,   0,   0,   0,   0,   0,   0,
	}, // ^^^^^^^^^^^^^^^^ White ^^^^^^^^^^^^^^^^
}

var bonusKnight = [2][64]int{
	{  // vvvvvvvvvvvvvvvv Black vvvvvvvvvvvvvvvv
	      -96, -33, -19, -12, -12, -19, -33, -96,
	      -26, -12,   0,   6,   6,   0, -12, -26,
	       -5,   6,  20,  27,  27,  20,   6,  -5,
	       -5,   6,  20,  27,  27,  20,   6,  -5,
	      -12,   0,  13,  20,  20,  13,   0, -12,
	      -26, -12,   0,   6,   6,   0, -12, -26,
	      -46, -33, -19, -12, -12, -19, -33, -46,
	      -67, -53, -40, -33, -33, -40, -53, -67,
	}, {
	      -52, -39, -27, -21, -21, -27, -39, -52,
	      -39, -27, -15,  -8,  -8, -15, -27, -39,
	      -27, -15,  -3,   2,   2,  -3, -15, -27,
	      -21,  -8,   2,   9,   9,   2,  -8, -21,
	      -21,  -8,   2,   9,   9,   2,  -8, -21,
	      -27, -15,  -3,   2,   2,  -3, -15, -27,
	      -39, -27, -15,  -8,  -8, -15, -27, -39,
	      -52, -39, -27, -21, -21, -27, -39, -52,
	}, // ^^^^^^^^^^^^^^^^ White ^^^^^^^^^^^^^^^^
}

var bonusBishop = [2][64]int{
	{  // vvvvvvvvvvvvvvvv Black vvvvvvvvvvvvvvvv
	       -8,  -8,  -6,  -4,  -4,  -6,  -8,  -8,
	       -8,   0,  -2,   0,   0,  -2,   0,  -8,
	       -6,  -2,   4,   2,   2,   4,  -2,  -6,
	       -4,   0,   2,   8,   8,   2,   0,  -4,
	       -4,   0,   2,   8,   8,   2,   0,  -4,
	       -6,  -2,   4,   2,   2,   4,  -2,  -6,
	       -8,   0,  -2,   0,   0,  -2,   0,  -8,
	      -20, -20, -17, -15, -15, -17, -20, -20,
	}, {
	      -29, -21, -17, -13, -13, -17, -21, -29,
	      -21, -13,  -9,  -5,  -5,  -9, -13, -21,
	      -17,  -9,  -5,  -2,  -2,  -5,  -9, -17,
	      -13,  -5,  -2,   2,   2,  -2,  -5, -13,
	      -13,  -5,  -2,   2,   2,  -2,  -5, -13,
	      -17,  -9,  -5,  -2,  -2,  -5,  -9, -17,
	      -21, -13,  -9,  -5,  -5,  -9, -13, -21,
	      -29, -21, -17, -13, -13, -17, -21, -29,
	}, // ^^^^^^^^^^^^^^^^ White ^^^^^^^^^^^^^^^^
}

var bonusRook = [2][64]int{
	{  // vvvvvvvvvvvvvvvv Black vvvvvvvvvvvvvvvv
	       -6,  -3,  -1,   1,   1,  -1,  -3,  -6,
	       -6,  -3,  -1,   1,   1,  -1,  -3,  -6,
	       -6,  -3,  -1,   1,   1,  -1,  -3,  -6,
	       -6,  -3,  -1,   1,   1,  -1,  -3,  -6,
	       -6,  -3,  -1,   1,   1,  -1,  -3,  -6,
	       -6,  -3,  -1,   1,   1,  -1,  -3,  -6,
	       -6,  -3,  -1,   1,   1,  -1,  -3,  -6,
	       -6,  -3,  -1,   1,   1,  -1,  -3,  -6,
	}, {
	        1,   1,   1,   1,   1,   1,   1,   1, 
	        1,   1,   1,   1,   1,   1,   1,   1,
	        1,   1,   1,   1,   1,   1,   1,   1,
	        1,   1,   1,   1,   1,   1,   1,   1,
	        1,   1,   1,   1,   1,   1,   1,   1,
	        1,   1,   1,   1,   1,   1,   1,   1,
	        1,   1,   1,   1,   1,   1,   1,   1,
	        1,   1,   1,   1,   1,   1,   1,   1, 
	}, // ^^^^^^^^^^^^^^^^ White ^^^^^^^^^^^^^^^^
}

var bonusQueen = [2][64]int{
	{  // vvvvvvvvvvvvvvvv Black vvvvvvvvvvvvvvvv
	        4,   4,   4,   4,   4,   4,   4,   4,
	        4,   4,   4,   4,   4,   4,   4,   4,
	        4,   4,   4,   4,   4,   4,   4,   4,
	        4,   4,   4,   4,   4,   4,   4,   4,
	        4,   4,   4,   4,   4,   4,   4,   4,
	        4,   4,   4,   4,   4,   4,   4,   4,
	        4,   4,   4,   4,   4,   4,   4,   4,
	        4,   4,   4,   4,   4,   4,   4,   4,
	}, {
	      -40, -27, -21, -15, -15, -21, -27, -40,
	      -27, -15,  -9,  -3,  -3,  -9, -15, -27,
	      -21,  -9,  -3,   3,   3,  -3,  -9, -21,
	      -15,  -3,   3,   9,   9,   3,  -3, -15,
	      -15,  -3,   3,   9,   9,   3,  -3, -15,
	      -21,  -9,  -3,   3,   3,  -3,  -9, -21,
	      -27, -15,  -9,  -3,  -3,  -9, -15, -27,
	      -40, -27, -21, -15, -15, -21, -27, -40,
	}, // ^^^^^^^^^^^^^^^^ White ^^^^^^^^^^^^^^^^
}

var bonusKing = [2][64]int{
	{  // vvvvvvvvvvvvvvvv Black vvvvvvvvvvvvvvvv
	       47,  59,  34,  10,  10,  34,  59,  47,
	       59,  71,  47,  23,  23,  47,  71,  59,
	       71,  83,  59,  34,  34,  59,  83,  71,
	       83,  95,  71,  47,  47,  71,  95,  83,
	       95, 107,  83,  59,  59,  83, 107,  95,
	      107, 119,  95,  71,  71,  95, 119, 107,
	      131, 143, 119,  95,  95, 119, 143, 131,
	      143, 155, 131, 107, 107, 131, 155, 143,
	}, {
		9,  38,  52,  67,  67,  52,  38,   9,
	       38,  67,  82,  96,  96,  82,  67,  38,
	       52,  82,  96, 111, 111,  96,  82,  52,
	       67,  96, 111, 125, 125, 111,  96,  67,
	       67,  96, 111, 125, 125, 111,  96,  67,
	       52,  82,  96, 111, 111,  96,  82,  52,
	       38,  67,  82,  96,  96,  82,  67,  38,
		9,  38,  52,  67,  67,  52,  38,   9,
	}, // ^^^^^^^^^^^^^^^^ White ^^^^^^^^^^^^^^^^
}

var bonusPassedPawn = [8]Score{
	{0, 0}, {0, 3}, {0, 7}, {17, 17}, {51, 35}, {102, 59}, {170, 91}, {0, 0},
}
var extraPassedPawn = [8]int{
	0, 0, 0, 1, 3, 6, 10, 0,
}

var extraKnight = [64]int{
     // vvvvvvvvvvv Black vvvvvvvvvvvv
	0,  0,  0,  0,  0,  0,  0,  0,
	0,  0,  0,  0,  0,  0,  0,  0,
	0,  2,  8,  8,  8,  8,  2,  0,
	0,  4, 13, 17, 17, 13,  4,  0,
	0,  2,  8, 13, 13,  8,  2,  0,
	0,  0,  2,  4,  4,  2,  0,  0,
	0,  0,  0,  0,  0,  0,  0,  0,
	0,  0,  0,  0,  0,  0,  0,  0,
     // ^^^^^^^^^^^ White ^^^^^^^^^^^^
}

var extraBishop = [64]int{
     // vvvvvvvvvvv Black vvvvvvvvvvvv
	0,  0,  0,  0,  0,  0,  0,  0,
	0,  0,  0,  0,  0,  0,  0,  0,
	0,  2,  4,  4,  4,  4,  2,  0,
	0,  5, 10, 10, 10, 10,  5,  0,
	0,  2,  5,  5,  5,  5,  2,  0,
	0,  0,  2,  2,  2,  2,  0,  0,
	0,  0,  0,  0,  0,  0,  0,  0,
	0,  0,  0,  0,  0,  0,  0,  0,
     // ^^^^^^^^^^^ White ^^^^^^^^^^^^
}

// [1] Pawn, [2] Knight, [3] Bishop, [4] Rook, [5] Queen
var bonusMinorThreat = [6]Score{
	{0, 0}, {3, 18}, {12, 24}, {12, 24}, {20, 50}, {20, 50},
}

// [1] Pawn, [2] Knight, [3] Bishop, [4] Rook, [5] Queen
var bonusMajorThreat = [6]Score{
	{0, 0}, {7, 18}, {7, 22}, {7, 22}, {7, 22}, {12, 24},
}

// [1] Pawn, [2] Knight, [3] Bishop, [4] Rook, [5] Queen
var bonusKingThreat = [6]int {
	0, 0, 2, 2, 3, 5,
}

// [1] Pawn, [2] Knight, [3] Bishop, [4] Rook, [5] Queen
var bonusCloseCheck = [6]int {
	0, 0, 0, 0, 8, 12,
}

// [1] Pawn, [2] Knight, [3] Bishop, [4] Rook, [5] Queen
var bonusDistanceCheck = [6]int {
	0, 0, 1, 1, 4, 6,
}

var kingSafety = [64]int {
	  0,   0,   1,   2,   3,   5,   7,  10,
	 13,  16,  20,  24,  29,  34,  39,  45,
	 51,  58,  65,  72,  80,  88,  97, 106,
	115, 125, 135, 146, 157, 168, 180, 192,
	205, 218, 231, 245, 259, 274, 289, 304,
	319, 334, 349, 364, 379, 394, 409, 424,
	439, 454, 469, 484, 499, 514, 529, 544,
	559, 574, 589, 604, 619, 634, 640, 640,
}

// Supported pawn bonus arranged from White point of view. The actual score
// uses the same values for midgame and endgame.
var bonusSupportedPawn = [64]int{
      // vvvvvvvvvvvv Black vvvvvvvvvvvv
	  0,  0,  0,  0,  0,  0,  0,  0,
	 62, 66, 66, 68, 68, 66, 66, 62,
	 31, 34, 34, 36, 36, 34, 34, 31,
	 13, 16, 16, 18, 18, 16, 16, 13,
	  4,  6,  6,  7,  7,  6,  6,  4,
	  1,  3,  3,  4,  4,  3,  3,  1,
	  0,  1,  1,  2,  2,  1,  1,  0,
	  0,  0,  0,  0,  0,  0,  0,  0,
     // ^^^^^^^^^^^^^ White ^^^^^^^^^^^^
}

// [1] Pawn, [2] Knight, [3] Bishop, [4] Rook, [5] Queen
var penaltyPawnThreat = [6]Score {
	{0, 0}, {0, 0}, {26, 35}, {26, 35}, {38, 49}, {43, 59},
}

// Penalty for doubled pawn: A to H, midgame/endgame.
var penaltyDoubledPawn = [8]Score{
	{7, 22}, {10, 24}, {12, 24}, {12, 24}, {12, 24}, {12, 24}, {10, 24}, {7, 22},
}

// Penalty for isolated pawn that is *not* exposed: A to H, midgame/endgame.
var penaltyIsolatedPawn = [8]Score{
	{12, 15}, {18, 17}, {20, 17}, {20, 17}, {20, 17}, {20, 17}, {18, 17}, {12, 15},
}

// Penalty for isolated pawn that is exposed: A to H, midgame/endgame.
var penaltyWeakIsolatedPawn = [8]Score{
	{18, 22}, {27, 26}, {30, 26}, {30, 26}, {30, 26}, {30, 26}, {27, 26}, {18, 22},
}

// Penalty for backward pawn that is *not* exposed: A to H, midgame/endgame.
var penaltyBackwardPawn = [8]Score{
	{10, 14}, {15, 16}, {17, 16}, {17, 16}, {17, 16}, {17, 16}, {15, 16}, {10, 14},
}

// Penalty for backward pawn that is exposed: A to H, midgame/endgame.
var penaltyWeakBackwardPawn = [8]Score{
	{15, 21}, {22, 23}, {25, 23}, {25, 23}, {25, 23}, {25, 23}, {22, 23}, {15, 21},
}

var mobilityKnight = [9]Score{
	{-32, -25}, {-21, -15}, {-4, -5}, {1, 0}, {7, 5}, {13, 10}, {18, 14}, {21, 15}, {22, 16},
}

var mobilityBishop = [16]Score{
	{-26, -23}, {-14, -11}, { 3,  0}, {10,  7}, {17, 14}, {24, 21}, {30, 27}, {34, 31},
	{ 37,  34}, { 38,  36}, {40, 37}, {41, 38}, {42, 39}, {43, 40}, {43, 40}, {43, 40},
}

var mobilityRook = [16]Score{
	{-23, -26}, {-15, -13}, {-2,  0}, { 0,  8}, { 3, 16}, { 6, 24}, { 9, 32}, {11, 40},
	{ 13,  48}, { 14,  54}, {15, 57}, {16, 59}, {17, 61}, {18, 61}, {18, 62},
}

var mobilityQueen = [16]Score{
	{-21, -20}, {-14, -12}, {-2, -3}, { 0,  0}, { 3,  5}, { 5,  9}, { 6, 14}, { 9, 19},
	{ 10,  20}, { 10,  20}, {11, 20}, {11, 20}, {11, 20}, {12, 20}, {12, 20}, {12, 20},
}

// Boxed rooks.
var kingBoxA = [2]Bitmask{
	bit[D1]|bit[C1]|bit[B1], bit[D8]|bit[C8]|bit[B8],
}

var kingBoxH = [2]Bitmask{
	bit[E1]|bit[F1]|bit[G1], bit[E8]|bit[F8]|bit[G8],
}

var rookBoxA = [2]Bitmask{
	bit[A1]|bit[B1]|bit[C1], bit[A8]|bit[B8]|bit[C8],
}

var rookBoxH = [2]Bitmask{
	bit[H1]|bit[G1]|bit[F1], bit[H8]|bit[G8]|bit[F8],
}
