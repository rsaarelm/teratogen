package gfx

// State wrapper structure for animation objects. Animations are not a proper
// part of the game world, they don't go into savegames, but they can be used
// to illustrate events in the world. Animations are run in separate
// goroutines, which communicate with the main game via the Anim object API.
type Anim struct {
	// Channel that receives an interval in nanoseconds since the last
	// update, causes the anim to draw itself.
	updateChan chan int64

	// The current graphics context, passed using the Update function.
	g Graphics

	// The position of the animation in the draw queue. Low z values are
	// drawn first.
	Z float64
}

func NewAnim(z float64) *Anim { return &Anim{make(chan int64), nil, z} }

// Update causes the anim to draw itself in the graphics context and advance
// it's state by elapsedNs. Calling with 0 elapsedNs will just make the anim
// redraw itself.
func (self *Anim) Update(g Graphics, elapsedNs int64) {
	self.g = g
	self.updateChan <- elapsedNs
	// Wait for the draw to finish
	<-self.updateChan
	// Just to make sure it isn't used inappropriately.
	self.g = nil
}

func (self *Anim) Close() {
	close(self.updateChan)
	<-self.updateChan
}

func (self *Anim) Closed() bool { return closed(self.updateChan) }

// StartDraw is called from the animation goroutine. It waits until the main
// engine wants the animation to draw itself. It returns the graphics context
// to draw into and the number of nanoseconds that the animation state should
// advance. Do not call this outside the animation goroutine.
func (self *Anim) StartDraw() (g Graphics, elapsedNs int64) {
	elapsedNs = <-self.updateChan
	g = self.g
	return
}

// StopDraw is called in from the animation goroutine when the animation has
// finished drawing its current state after calling StartDraw.
func (self *Anim) StopDraw() { self.updateChan <- 0 }
