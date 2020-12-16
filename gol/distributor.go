package gol

import (
	
	"fmt"
	"net/rpc"
	"strconv"

	"uk.ac.bris.cs/gameoflife/util"
)

const alive = 255
const dead = 0

var CalaulateHandler = "DistributedOperations.Calculate"
var AliveHandler = "DistributedOperations.AliveCells"

type Resource struct {
	World   [][]byte
	Turns   int
	Threads int
	Width   int
	Height  int
}

type ResponseCal struct {
	World	[][]byte
	X 		[]int
	Y 		[]int
	Turn 	[]int
}

type ResponseAlive struct {
	Alivecells []util.Cell
}


type distributorChannels struct {
	events    chan<- Event
	ioCommand chan<- ioCommand
	ioIdle    <-chan bool
	output    chan<- uint8
	input     <-chan uint8
	filename  chan string
}

func makeMatrix(height, width int) [][]byte {
	matrix := make([][]byte, height)
	for i := range matrix {
		matrix[i] = make([]byte, width)
	}
	return matrix
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, keyPresses <-chan rune) {


	Client, _ := rpc.Dial("tcp", "100.25.21.156:8030")
	defer Client.Close()

	c.ioCommand <- ioInput
	c.filename <- fmt.Sprint(strconv.Itoa(p.ImageWidth), "x", strconv.Itoa(p.ImageHeight))

	World := makeMatrix(p.ImageHeight, p.ImageWidth)
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			cellmessage := <-c.input
			World[y][x] = cellmessage
			if World[y][x] == alive {
				cell := CellFlipped{CompletedTurns: 0, Cell: util.Cell{X: x, Y: y}}
				c.events <- cell
			}
		}
	}
	
	resource := Resource{
		World:   World,
		Turns:   p.Turns,
		Threads: p.Threads,
		Width:   p.ImageWidth,
		Height:  p.ImageHeight,
	}
	
	//for turn := 0; turn <= p.Turns; turn++ {
		
		request := new(ResponseCal)
		
		Client.Call(CalaulateHandler, resource, request)
		
		
		
		X := request.X
		Y := request.Y 
		turns := request.Turn
		for i := 0; i<len(X); i++ {
			if (X[i]==-1||Y[i]==-1){
				c.events <- TurnComplete{CompletedTurns:turns[i]}
			}else {
				c.events <- CellFlipped{CompletedTurns:turns[i],Cell:util.Cell{X:X[i], Y:Y[i]}}
			}
		}
		
		
		resource.World = request.World
	//}
	
	
	// resource.World = request.World
	response := new(ResponseAlive)
	Client.Call(AliveHandler, resource, response)
	
	c.events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: response.Alivecells}

	c.ioCommand <- ioOutput
	c.filename <- fmt.Sprint(strconv.Itoa(p.ImageWidth), "x", strconv.Itoa(p.ImageHeight), "x", strconv.Itoa(p.Turns))
	
	
	finalWorld := resource.World
	for y := 0; y < resource.Height; y++ {
		for x := 0; x < resource.Width; x++ {
			c.output <- finalWorld[y][x]
		}
	}

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

// func mod(x, m int) int {
// 	return (x + m) % m
// }

// func calculateNeighbours(p Params, x, y int, world [][]byte) int {
// 	neighbours := 0
// 	for i := -1; i <= 1; i++ {
// 		for j := -1; j <= 1; j++ {
// 			if i != 0 || j != 0 {
// 				if world[mod(y+i, p.ImageHeight)][mod(x+j, p.ImageWidth)] == alive {
// 					neighbours++
// 				}
// 			}
// 		}
// 	}
// 	return neighbours
// }

// func calculateNextState(a int, b int, width int, p Params, world [][]byte, c distributorChannels, turn int) [][]byte {

// 	newWorld := make([][]byte, b-a)
// 	for i := range newWorld {
// 		newWorld[i] = make([]byte, width)
// 	}
// 	for y := 0; y < b-a; y++ {
// 		for x := 0; x < width; x++ {
// 			neighbours := calculateNeighbours(p, x, y+a, world)
// 			if world[a+y][x] == alive {
// 				if neighbours == 2 || neighbours == 3 {
// 					newWorld[y][x] = alive

// 				} else {
// 					newWorld[y][x] = dead
// 					c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: x, Y: y + a}}
// 				}
// 			} else {
// 				if neighbours == 3 {
// 					newWorld[y][x] = alive

// 					c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: x, Y: y + a}}
// 				} else {
// 					newWorld[y][x] = dead
// 				}
// 			}
// 		}
// 	}
// 	return newWorld
// }

// func makeMatrix(height, width int) [][]byte {
// 	matrix := make([][]byte, height)
// 	for i := range matrix {
// 		matrix[i] = make([]byte, width)
// 	}
// 	return matrix
// }

// func worker(a int, b int, width int, p Params, world [][]byte, out chan [][]byte, c distributorChannels, turn int) {
// 	imagePart := calculateNextState(a, b, width, p, world, c, turn)

// 	out <- imagePart
// }

// func aliveCells(p Params, world [][]byte) []util.Cell {
// 	alivecells := make([]util.Cell, 1)
// 	for y := 0; y < p.ImageHeight; y++ {
// 		for x := 0; x < p.ImageWidth; x++ {

// 			if world[y][x] == alive {
// 				cell := util.Cell{X: x, Y: y}
// 				alivecells = append(alivecells, cell)
// 			}
// 		}
// 	}

// 	alivecells = alivecells[1:]
// 	return alivecells
// }

// func keypress(c distributorChannels, param Params, keyPresses <-chan rune, turn int, World [][]byte) {

// 	select {
// 	case key := <-keyPresses:
// 		switch key {
// 		case 's':
// 			c.ioCommand <- ioOutput
// 			c.filename <- fmt.Sprint(strconv.Itoa(param.ImageWidth), "x", strconv.Itoa(param.ImageHeight), "x", strconv.Itoa(param.Turns))

// 			for y := 0; y < param.ImageHeight; y++ {
// 				for x := 0; x < param.ImageWidth; x++ {
// 					c.output <- World[y][x]
// 				}
// 			}
// 			c.events <- StateChange{CompletedTurns: turn, NewState: Executing}

// 		case 'q':
// 			c.ioCommand <- ioOutput
// 			c.filename <- fmt.Sprint(strconv.Itoa(param.ImageWidth), "x", strconv.Itoa(param.ImageHeight), "x", strconv.Itoa(param.Turns))

// 			for y := 0; y < param.ImageHeight; y++ {
// 				for x := 0; x < param.ImageWidth; x++ {
// 					c.output <- World[y][x]
// 				}
// 			}
// 			c.events <- StateChange{CompletedTurns: turn, NewState: Quitting}
// 			close(c.events)
// 		case 'p':
// 			c.events <- StateChange{CompletedTurns: turn, NewState: Paused}
// 			fmt.Println("Current turn:", turn)
// 			keyp := <-keyPresses
// 			for {

// 				if keyp == 'p' {
// 					c.events <- StateChange{CompletedTurns: turn, NewState: Executing}
// 					fmt.Println("Continuing")
// 					break
// 				} else {
// 					fmt.Println("Pause")
// 					continue
// 				}

// 			}
// 		}

// 	default:
// 		{
// 		}
// 	}

// }

// workerHeight := p.ImageHeight / p.Threads

// turn := 0
// ticker := time.NewTicker(2 * time.Second)

// for t := 0; t < p.Turns; t++ {

// 	out := make([]chan [][]byte, p.Threads)
// 	for i := 0; i < p.Threads; i++ {
// 		out[i] = make(chan [][]byte)
// 	}

// 	select {
// 	case <-ticker.C:
// 		c.events <- AliveCellsCount{
// 			CompletedTurns: turn,
// 			CellsCount:     len(aliveCells(p, World)),
// 		}
// 	default:
// 		{
// 		}
// 	}
// 	for i := 0; i < p.Threads; i++ {
// 		if i <= p.Threads-2 {
// 			go worker(i*workerHeight, (i+1)*workerHeight, p.ImageWidth, p, World, out[i], c, turn)
// 		} else {
// 			go worker(i*workerHeight, p.ImageHeight, p.ImageWidth, p, World, out[i], c, turn)
// 		}

// 	}

// 	newWorld := makeMatrix(0, 0)

// 	for i := 0; i < p.Threads; i++ {
// 		part := <-out[i]

// 		newWorld = append(newWorld, part...)

// 	}
// 	World = newWorld
// 	c.events <- TurnComplete{CompletedTurns: turn}

// 	turn++

// 	keypress(c, p, keyPresses, t, World)

// }
// ticker.Stop()

// alivecells := aliveCells(p, World)
// c.events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: alivecells}
