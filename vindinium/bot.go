package vindinium

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
)

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

type CashBot struct {
	LifeBoard      [][]int
	DirectionBoard [][]Direction
	Size           int
	MaxLife        int
	GoDirection    Direction
	survivalMode   bool
}

func (b *CashBot) Move(state *State) Direction {
	board := state.Game.Board
	board.parseTiles()

	b.buildBoards(state)
	//b.fillBoards(state)

//	b.PrintDirectionBoard()
//	fmt.Println("Going ", b.GoDirection)

	hero := state.Hero
	b.survivalMode = false

	lazyMode := true

	for badHeroIndex := range state.Game.Heroes {
		badHero := state.Game.Heroes[badHeroIndex]
		if 	badHero.MineCount > hero.MineCount ||
			( badHero.MineCount == hero.MineCount && badHero.Gold > hero.Gold )  {
			lazyMode = false
		}
	}

	if lazyMode && hero.Life > 90 {
		return "Stay"
	}

	if hero.Life < 50 || lazyMode{
		b.survivalMode = true
	}

	

	des := b.BoardWalker(state)
	b.PrintDirectionBoard()

	return des.Dir
}

type Destiny struct {
	Pos Position
	Dir Direction
}

func (b *CashBot) BoardWalker (state *State)  Destiny {

	hero := state.Hero

	from := make(chan Destiny, 10000)
	from <- Destiny{Pos: *hero.Pos, Dir: "Stay"}

	seen := make(map[Position]bool)
	board := state.Game.Board
	size := state.Game.Board.Size

	searchlevel := 0
	for des := range from {

		searchlevel++
		if searchlevel > 1000 {
			return Destiny{Pos: *state.Hero.Pos, Dir: randDir()}
		}
		

		if 	seen[des.Pos] ||
			des.Pos.X < 0 ||
			des.Pos.Y < 0 ||
			des.Pos.X >= size ||
			des.Pos.Y >= size {

			if len(from) <= 0 {
				break
			}

			continue
		}
			seen[des.Pos] = true
			tile := board.Tileset[des.Pos.X][des.Pos.Y]

			if b.survivalMode && tile == TAVERN {
				return des
			}

			if reflect.TypeOf(tile).String() == "*vindinium.HeroTile" {
				heroTile := tile.(*HeroTile)
				if(heroTile.Id != state.Hero.Id){
					hero := state.Game.Heroes[heroTile.Id-1]
					if(hero.Life < state.Hero.Life){
						fmt.Printf("Found Hero %d with less life!", heroTile.Id)
						return des
					}
				}
			}

			if ! b.survivalMode && reflect.TypeOf(tile).String() == "*vindinium.MineTile" {
				mine := tile.(*MineTile)
				mineId, _ := strconv.Atoi(mine.HeroId)
				if mineId != state.Hero.Id {
					return des
				}
			}

		
			if tile == AIR  || reflect.TypeOf(tile).String() == "*vindinium.HeroTile"  {

				if reflect.TypeOf(tile).String() == "*vindinium.HeroTile" {
					heroTile := tile.(*HeroTile)
					if(heroTile.Id != state.Hero.Id){
						continue;
					}
				}

				
				b.DirectionBoard[des.Pos.X][des.Pos.Y] = des.Dir[:1]
				if des.Dir == "Stay" {
					from <- Destiny{Pos: Position{X: des.Pos.X - 1, Y: des.Pos.Y    }, Dir: "North"}
					from <- Destiny{Pos: Position{X: des.Pos.X    , Y: des.Pos.Y - 1}, Dir: "West"}
					from <- Destiny{Pos: Position{X: des.Pos.X + 1, Y: des.Pos.Y    }, Dir: "South"}
					from <- Destiny{Pos: Position{X: des.Pos.X    , Y: des.Pos.Y + 1}, Dir: "East"}
				} else {
					from <- Destiny{Pos: Position{X: des.Pos.X - 1, Y: des.Pos.Y}, Dir: des.Dir}
					from <- Destiny{Pos: Position{X: des.Pos.X, Y: des.Pos.Y - 1}, Dir: des.Dir}
					from <- Destiny{Pos: Position{X: des.Pos.X + 1, Y: des.Pos.Y   }, Dir: des.Dir}
					from <- Destiny{Pos: Position{X: des.Pos.X    , Y: des.Pos.Y + 1}, Dir: des.Dir}
				}
			}
	
		
	
	}
	return Destiny{Pos: *state.Hero.Pos, Dir: "Stay"}
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

func (b *CashBot) fillBoards(state *State) {
	hero := state.Hero
	b.survivalMode = false
	if hero.Life < 70 {
		b.survivalMode = true
	}
	b.fillBoardsRecursive(state, hero.Life, hero.Pos, "Stay")
	//b.PrintLifeBoard(nil)
}

func (b *CashBot) fillBoardsRecursive(state *State, life int, pos *Position, way Direction) {

	if pos.X < 0 ||
		pos.Y < 0 ||
		pos.X >= b.Size-1 ||
		pos.Y >= b.Size-1 ||
		life < 0 || 
		(!b.survivalMode && life < 70) {
		return
	}

	board := state.Game.Board
	tile := board.Tileset[pos.X][pos.Y]

	if b.survivalMode && tile == TAVERN {
		if life > b.MaxLife {
			b.MaxLife = life
			b.GoDirection = way
		}
		return
	}

			if !b.survivalMode && reflect.TypeOf(tile).String() == "*vindinium.MineTile" {
				mine := tile.(*MineTile)
				mineId, _ := strconv.Atoi(mine.HeroId)
				if life > b.MaxLife && mineId != state.Hero.Id {
					/*
						fmt.Println("Found a mine!",mine.HeroId)
						fmt.Println("My id ", string(state.Hero.Id) )
						fmt.Println("at pos", pos)
						fmt.Println("with life", life)
					*/
					b.MaxLife = life
					b.GoDirection = way
					//fmt.Println("Let's go ",way)
				}
				return
			}

	if tile == AIR || reflect.TypeOf(tile).String() == "*vindinium.HeroTile" {

		if life > b.LifeBoard[pos.X][pos.Y] {
			b.LifeBoard[pos.X][pos.Y] = life
			b.DirectionBoard[pos.X][pos.Y] = way[:1]


			if way == "Stay" {
				b.fillBoardsRecursive(state, life-1, &Position{pos.X, pos.Y - 1}, "West")
				b.fillBoardsRecursive(state, life-1, &Position{pos.X - 1, pos.Y}, "North")
				b.fillBoardsRecursive(state, life-1, &Position{pos.X, pos.Y + 1}, "East")
				b.fillBoardsRecursive(state, life-1, &Position{pos.X + 1, pos.Y}, "South")
			} else {
				b.fillBoardsRecursive(state, life-1, &Position{pos.X, pos.Y - 1}, way)
				b.fillBoardsRecursive(state, life-1, &Position{pos.X - 1, pos.Y}, way)
				b.fillBoardsRecursive(state, life-1, &Position{pos.X, pos.Y + 1}, way)
				b.fillBoardsRecursive(state, life-1, &Position{pos.X + 1, pos.Y}, way)
			}

		}
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
