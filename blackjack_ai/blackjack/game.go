package blackjack

import (
	"errors"
	"fmt"

	"github.com/jeremy-miller/gophercises/deck"
)

type state uint8

const (
	stateBet state = iota
	statePlayerTurn
	stateDealerTurn
	stateHandOver
)

type Options struct {
	Decks           int
	Hands           int
	BlackjackPayout float64
}

type Game struct {
	// unexported fields
	numDecks        int
	numHands        int
	blackjackPayout float64

	state state
	deck  []deck.Card

	player    []deck.Card
	playerBet int
	balance   int

	dealer   []deck.Card
	dealerAI AI
}

func New(opts Options) Game {
	g := Game{
		dealerAI: dealerAI{},
		balance:  0,
	}
	if opts.Decks == 0 {
		opts.Decks = 3
	}
	if opts.Hands == 0 {
		opts.Hands = 100
	}
	if opts.BlackjackPayout == 0 {
		opts.BlackjackPayout = 1.5
	}
	g.numDecks = opts.Decks
	g.numHands = opts.Hands
	g.blackjackPayout = opts.BlackjackPayout
	return g
}

func (g *Game) Play(ai AI) int {
	g.deck = nil
	min := 52 * g.numDecks / 3 // reshuffle when we're down to 1/3 of total cards left in deck
	for i := 0; i < g.numHands; i++ {
		shuffled := false
		if len(g.deck) < min {
			g.deck = deck.New(deck.Deck(g.numDecks), deck.Shuffle)
			shuffled = true
		}
		bet(g, ai, shuffled)
		deal(g)
		if Blackjack(g.dealer...) {
			endHand(g, ai)
			continue
		}
		for g.state == statePlayerTurn {
			hand := make([]deck.Card, len(g.player))
			copy(hand, g.player)
			move := ai.Play(hand, g.dealer[0]) // only pass in first card of dealer's hand
			err := move(g)
			switch err {
			case errBust:
				_ := MoveStand(g) // ignore error since it always returns nil
			case nil:
				// noop
			default:
				panic(err)
			}
		}
		for g.state == stateDealerTurn {
			hand := make([]deck.Card, len(g.dealer))
			copy(hand, g.dealer)
			move := g.dealerAI.Play(hand, g.dealer[0])
			move(g)
		}
		endHand(g, ai)
	}
	return g.balance
}

func bet(g *Game, ai AI, shuffled bool) {
	g.playerBet = ai.Bet(shuffled)
}

func deal(g *Game) {
	g.player = make([]deck.Card, 0, 5) // likely won't have more than 5 cards in hand in a game
	g.dealer = make([]deck.Card, 0, 5)
	var card deck.Card
	for i := 0; i < 2; i++ {
		card, g.deck = draw(g.deck)
		g.player = append(g.player, card)
		card, g.deck = draw(g.deck)
		g.dealer = append(g.dealer, card)
	}
	g.state = statePlayerTurn
}

// Score will take in a hand of cards and return the best blackjack score possible with that hand.
func Score(hand ...deck.Card) int { // using variadic so user can pass in just one card if desired
	minScore := minScore(hand...)
	if minScore > 11 {
		return minScore
	}
	for _, c := range hand {
		if c.Rank == deck.Ace {
			return minScore + 10 // ace is currently worth 1, so adding 10 to make it worth 11
		}
		// only 1 ace can be used as 11 and still be under 21
	}
	return minScore
}

func minScore(hand ...deck.Card) int {
	score := 0
	for _, c := range hand {
		score += min(int(c.Rank), 10)
	}
	return score
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Soft returns true if the score of a hand is a soft score (i.e. if an ace is being counted as 11 points).
func Soft(hand ...deck.Card) bool {
	minScore := minScore(hand...)
	score := Score(hand...)
	return minScore != score
}

// Blackjack returns true if a hand is a blackjack
func Blackjack(hand ...deck.Card) bool {
	return len(hand) == 2 && Score(hand...) == 21
}

var (
	errBust = errors.New("hand score exceeded 21")
)

type Move func(*Game) error

func MoveHit(g *Game) error {
	hand := g.currentHand()
	var card deck.Card
	card, g.deck = draw(g.deck)
	*hand = append(*hand, card)
	if Score(*hand...) > 21 {
		return errBust
	}
	return nil
}

func (g *Game) currentHand() *[]deck.Card {
	switch g.state {
	case statePlayerTurn:
		return &g.player
	case stateDealerTurn:
		return &g.dealer
	default: // shouldn't ever happen, if so, there's a bug
		panic("it isn't currently any player's turn.")
	}
}

func draw(cards []deck.Card) (deck.Card, []deck.Card) {
	return cards[0], cards[1:]
}

func MoveDouble(g *Game) error {
	if len(g.player) != 2 {
		return errors.New("can only double on a hand with 2 cards")
	}
	g.playerBet *= 2
	_ := MoveHit(g) // ignore error since only error can be errBust, so we'll have to stand anyways
	return MoveStand(g)
}

func MoveStand(g *Game) error {
	g.state++
	return nil
}

func endHand(g *Game, ai AI) {
	pScore, dScore := Score(g.player...), Score(g.dealer...)
	pBlackjack, dBlackjack := Blackjack(g.player...), Blackjack(g.dealer...)
	winnings := g.playerBet
	switch {
	case pBlackjack && dBlackjack:
		winnings = 0
	case dBlackjack:
		winnings = -winnings
	case pBlackjack:
		winnings = int(float64(winnings) * g.blackjackPayout)
	case pScore > 21:
		winnings = -winnings
	case dScore > 21:
		// win
	case pScore > dScore:
		// win
	case dScore > pScore:
		winnings = -winnings
	case pScore == dScore:
		winnings = 0
	}
	g.balance += winnings
	fmt.Println()
	ai.Results([][]deck.Card{g.player}, g.dealer)
	g.player = nil // clear out hands
	g.dealer = nil
}
