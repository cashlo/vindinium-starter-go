package vindinium

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Direction string

var DIRS = []Direction{"Stay", "North", "South", "East", "West"}

const LifeToPanic = 50

func randDir() Direction {
	dir := DIRS[rand.Intn(len(DIRS))]

	return dir
}

type Bot interface {
	Move(state *State) Direction
}

type RandomBot struct{}

func (b *RandomBot) Move(state *State) Direction {
	return randDir()
}

type FighterBot struct{}

func (b *FighterBot) Move(state *State) Direction {
	println(state)
	// g := NewGame(state)
	// Do something awesome
	return randDir()
}

type CashBot struct {
	LifeBoard      [][]int
	DirectionBoard [][]Direction
	Size           int
	MaxLife        int
	GoDirection    Direction
	survivalMode   bool

	posToVisit chan Position
	posSeen    chan bool
}

func (b *CashBot) Move(state *State) Direction {
	board := state.Game.Board
	board.parseTiles()

	hero := state.Hero
	b.survivalMode = false

	lazyMode := true

	fmt.Println("My ID: ", hero.Id, "My mines: ", hero.MineCount, " My gold: ", hero.Gold)
	for badHeroIndex := range state.Game.Heroes {
		badHero := state.Game.Heroes[badHeroIndex]
		if hero.Id == badHero.Id {
			continue
		}
		fmt.Println("their mines: ", badHero.MineCount, " Their gold: ", badHero.Gold)
		if badHero.MineCount > hero.MineCount ||
			(badHero.MineCount == hero.MineCount && badHero.Gold >= hero.Gold) {
			lazyMode = false
		}
	}

	if lazyMode && hero.Life > 90 {
		return "Stay"
	}

	if hero.Life < LifeToPanic || lazyMode {
		b.survivalMode = true
	}

	out := make(chan Destiny)
	go b.BoardWalker(state, out)
	go b.oneSec(out)

	des := <-out
	return des.Dir
}

type Destiny struct {
	Pos Position
	Dir Direction
}

func (b *CashBot) oneSec(out chan Destiny) {
	time.Sleep(1 * time.Second)
	out <- Destiny{Pos: Position{0, 0}, Dir: randDir()}
	return
}

func (b *CashBot) seenCoordinator(posIn <-chan Position, seenOut chan<- bool) {
	seen := make(map[Position]bool)
	for pos := range posIn {
		seenOut <- seen[pos]
		seen[pos] = true
	}
	return
}

func (b *CashBot) BoardWalker(state *State, out chan Destiny) {

	b.posToVisit = make(chan Position)
	b.posSeen = make(chan bool)
	go b.seenCoordinator(b.posToVisit, b.posSeen)

	hero := state.Hero
	go b.visitNode(
		state,
		Destiny{Pos: Position{X: hero.Pos.X - 1, Y: hero.Pos.Y}, Dir: "North"},
		hero.Life,
		out)

	go b.visitNode(
		state,
		Destiny{Pos: Position{X: hero.Pos.X, Y: hero.Pos.Y - 1}, Dir: "West"},
		hero.Life,
		out)

	go b.visitNode(
		state,
		Destiny{Pos: Position{X: hero.Pos.X + 1, Y: hero.Pos.Y}, Dir: "South"},
		hero.Life,
		out)

	go b.visitNode(
		state,
		Destiny{Pos: Position{X: hero.Pos.X, Y: hero.Pos.Y + 1}, Dir: "East"},
		hero.Life,
		out)
	return
}

func (b *CashBot) visitNode(state *State, des Destiny, life int, out chan Destiny) {

	board := state.Game.Board
	size := state.Game.Board.Size

	b.posToVisit <- des.Pos
	if <-b.posSeen ||
		des.Pos.X < 0 ||
		des.Pos.Y < 0 ||
		des.Pos.X >= size ||
		des.Pos.Y >= size ||
		(!b.survivalMode && life < LifeToPanic) {
		return
	}

	tile := board.Tileset[des.Pos.X][des.Pos.Y]

	if b.survivalMode && tile == TAVERN {
		out <- des
		return
	}

	if reflect.TypeOf(tile).String() == "*vindinium.HeroTile" {
		heroTile := tile.(*HeroTile)
		if heroTile.Id != state.Hero.Id {
			hero := state.Game.Heroes[heroTile.Id-1]
			if hero.Life < state.Hero.Life && hero.MineCount > 1 {
				fmt.Printf("Found Hero %d with less life!\n", heroTile.Id)
				out <- des
				return
			} else if hero.Life < state.Hero.Life && state.Hero.MineCount > 0 {
				fmt.Println("Run for my life!!")
				b.survivalMode = true
				
			}
		}
	}

	if !b.survivalMode && reflect.TypeOf(tile).String() == "*vindinium.MineTile" {
		mine := tile.(*MineTile)
		mineId, _ := strconv.Atoi(mine.HeroId)
		if mineId != state.Hero.Id {
			out <- des
			return
		}
	}

	if tile == AIR || reflect.TypeOf(tile).String() == "*vindinium.HeroTile" {

		if reflect.TypeOf(tile).String() == "*vindinium.HeroTile" {
			heroTile := tile.(*HeroTile)
			if heroTile.Id != state.Hero.Id {
				return
			}
		}

		go b.visitNode(
			state,
			Destiny{Pos: Position{X: des.Pos.X - 1, Y: des.Pos.Y}, Dir: des.Dir},
			life-1,
			out)

		go b.visitNode(
			state,
			Destiny{Pos: Position{X: des.Pos.X, Y: des.Pos.Y - 1}, Dir: des.Dir},
			life-1,
			out)

		go b.visitNode(
			state,
			Destiny{Pos: Position{X: des.Pos.X + 1, Y: des.Pos.Y}, Dir: des.Dir},
			life-1,
			out)

		go b.visitNode(
			state,
			Destiny{Pos: Position{X: des.Pos.X, Y: des.Pos.Y + 1}, Dir: des.Dir},
			life-1,
			out)
	}
}

func (b *CashBot) buildBoards(state *State) {
	b.Size = state.Game.Board.Size
	fmt.Println("Size is ", b.Size)
	b.MaxLife = 0
	b.LifeBoard = make([][]int, b.Size)
	b.DirectionBoard = make([][]Direction, b.Size)
	for i := 0; i < b.Size; i++ {
		b.LifeBoard[i] = make([]int, b.Size)
		b.DirectionBoard[i] = make([]Direction, b.Size)
	}
}

func (b *CashBot) PrintDirectionBoard() {

	size := b.Size
	fmt.Println(strings.Repeat("=", size))
	for i := range b.DirectionBoard {
		for j := range b.DirectionBoard[i] {
			fmt.Printf("%1s", b.DirectionBoard[i][j])
		}
		println()
	}
}

func (b *CashBot) PrintLifeBoard(pos *Position) {

	size := b.Size
	fmt.Println(strings.Repeat("=", size))
	for i := range b.LifeBoard {
		for j := range b.LifeBoard[i] {
			if pos != nil && pos.X == i && pos.Y == j {
				fmt.Printf("xxx")
			} else {
				fmt.Printf("%3d", b.LifeBoard[i][j])
			}

		}
		println()
	}
}
