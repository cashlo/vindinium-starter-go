package vindinium

import "math/rand"
import "fmt"

type Direction string

var DIRS = []Direction{"Stay", "North", "South", "East", "West"}

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

type CashBot struct{
	LifeBoard	[][]int;
	DirectionBoard	[][]Direction;
	
}

func (b *CashBot) Move(state *State) Direction {
	board := state.Game.Board
	board.parseTiles()
	
	b.buildBoards(state)
	b.fillBoards(state)

	return "South"
}

func (b *CashBot) buildBoards(state *State) {
	size := state.Game.Board.Size;

	b.LifeBoard = make([][]int, size)
	b.DirectionBoard = make([][]Direction, size)
	for i := 0; i < size; i++ {
		b.LifeBoard[i] = make([]int, size)
		b.DirectionBoard[i] = make([]Direction, size)
	}
	
}

func (b *CashBot) fillBoards(state *State) {
	hero := state.Hero
	fmt.Printf("%v\n", hero)
	b.fillBoardsRecursive(state, hero.Life, hero.Pos)
}

func (b *CashBot) fillBoardsRecursive(state *State, life int, pos *Position){
	board := state.Game.Board

	if board.Tileset[pos.X][pos.Y] == AIR && life > b.LifeBoard[pos.X][pos.Y]{
		b.LifeBoard[pos.X][pos.Y] = life
	}

}
