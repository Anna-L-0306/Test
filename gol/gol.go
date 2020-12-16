package gol



// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {

	
	
	ioCommand := make(chan ioCommand)

	ioIdle := make(chan bool)

	
	fileChannel := make(chan string)
	inputChannel := make(chan uint8)
	outputChannel := make(chan uint8)

	distributorChannels := distributorChannels{
		events,
		ioCommand,
		ioIdle,
		outputChannel,
		inputChannel,
		fileChannel,
	}

	go distributor(p, distributorChannels, keyPresses)

	ioChannels := ioChannels{

		command:  ioCommand,
		idle:     ioIdle,
		filename: fileChannel,
		output:   outputChannel,
		input:    inputChannel,
	}

	go startIo(p, ioChannels)

}
