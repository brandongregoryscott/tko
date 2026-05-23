package grid

import (
	"fmt"
	"net"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

// resolveLocalPort opens a UDP listener on an ephemeral port and returns the port number.
func (c *Controller) resolveLocalPort() (int, error) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return 0, err
	}
	port := conn.LocalAddr().(*net.UDPAddr).Port
	conn.Close()
	return port, nil
}

// discover performs the serialosc handshake to find a grid device.
func (c *Controller) discover() {
	c.mu.Lock()
	localPort := c.localPort
	c.mu.Unlock()

	debug.Printf("[grid] discover starting on port %d", localPort)

	// Open a UDP connection for discovery.
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		debug.Printf("[grid] discover: resolve error: %v", err)
		c.sendMsg(GridErrorMsg{Err: fmt.Errorf("grid discovery: %w", err)})
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		debug.Printf("[grid] discover: listen error: %v", err)
		c.sendMsg(GridErrorMsg{Err: fmt.Errorf("grid discovery: %w", err)})
		return
	}
	defer conn.Close()

	// Send /serialosc/list to the serialosc daemon (port 12002).
	msg, err := buildSerialoscList("127.0.0.1", localPort)
	if err != nil {
		debug.Printf("[grid] discover: build message error: %v", err)
		c.sendMsg(GridErrorMsg{Err: fmt.Errorf("grid discovery: %w", err)})
		return
	}
	serialoscAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12002")
	if err != nil {
		debug.Printf("[grid] discover: serialosc addr error: %v", err)
		c.sendMsg(GridErrorMsg{Err: fmt.Errorf("grid discovery: %w", err)})
		return
	}
	if _, err := conn.WriteTo(msg, serialoscAddr); err != nil {
		debug.Printf("[grid] discover: send error: %v", err)
		c.sendMsg(GridErrorMsg{Err: fmt.Errorf("grid discovery: %w", err)})
		return
	}
	debug.Printf("[grid] discover: sent /serialosc/list, waiting for response...")

	// Wait for /serialosc/device response with a 2-second timeout.
	buf := make([]byte, 65535)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				debug.Printf("[grid] discover: timeout — no serialosc response")
				c.mu.Lock()
				c.state = Disconnected
				c.mu.Unlock()
				return
			}
			select {
			case <-c.done:
				return
			default:
			}
			debug.Printf("[grid] discover: read error: %v", err)
			c.sendMsg(GridErrorMsg{Err: fmt.Errorf("grid discovery: %w", err)})
			return
		}

		id, devType, port := parseSerialoscDevice(buf[:n])
		if port == 0 {
			debug.Printf("[grid] discover: unexpected OSC response, not a device")
			continue
		}

		debug.Printf("[grid] discover: found device id=%s type=%s port=%d", id, devType, port)

		// Close the discovery connection before starting the listener,
		// since listenForGridEvents needs to bind the same local port.
		conn.Close()

		c.mu.Lock()
		c.info = &GridInfo{ID: id, Type: devType, Port: port}
		c.gridClient = osc.NewClient("127.0.0.1", port)
		c.state = Connected
		c.mu.Unlock()

		// Configure the device so it sends key events to us and accepts
		// our LED messages with a known prefix.
		c.configureDevice(localPort)

		// Start listening for grid key events.
		go c.listenForGridEvents()

		c.sendMsg(GridConnectedMsg{Info: GridInfo{ID: id, Type: devType, Port: port}})
		return
	}
}

// oscPrefix is the OSC address prefix used for all grid messages.
const oscPrefix = "/tko"

// configureDevice tells the grid device where to send key events and
// sets the OSC prefix so all messages use a consistent namespace.
func (c *Controller) configureDevice(localPort int) {
	c.mu.Lock()
	client := c.gridClient
	c.mu.Unlock()
	if client == nil {
		debug.Printf("[grid] configureDevice: no client")
		return
	}

	debug.Printf("[grid] configureDevice: setting prefix=%s host=127.0.0.1 port=%d", oscPrefix, localPort)

	client.Send(osc.NewMessage("/sys/prefix", oscPrefix))
	client.Send(osc.NewMessage("/sys/host", "127.0.0.1"))
	client.Send(osc.NewMessage("/sys/port", int32(localPort)))
}

// listenForGridEvents starts the go-osc server that receives /grid/key
// events from the hardware and forwards them as GridKeyMsg.
func (c *Controller) listenForGridEvents() {
	c.mu.Lock()
	addr := fmt.Sprintf("127.0.0.1:%d", c.localPort)
	c.mu.Unlock()

	debug.Printf("[grid] listenForGridEvents: binding to %s", addr)

	server := &osc.Server{Addr: addr}
	dispatcher := osc.NewStandardDispatcher()
	dispatcher.AddMsgHandler(oscPrefix+"/grid/key", func(msg *osc.Message) {
		if len(msg.Arguments) >= 3 {
			x, _ := msg.Arguments[0].(int32)
			y, _ := msg.Arguments[1].(int32)
			s, _ := msg.Arguments[2].(int32)
			c.sendMsg(GridKeyMsg{X: int(x), Y: int(y), State: int(s)})
		}
	})
	server.Dispatcher = dispatcher

	c.mu.Lock()
	c.gridServer = server
	c.mu.Unlock()

	// ListenAndServe blocks until CloseConnection is called.
	if err := server.ListenAndServe(); err != nil {
		select {
		case <-c.done:
			return
		default:
		}
		debug.Printf("[grid] listenForGridEvents: server error: %v", err)
		c.sendMsg(GridErrorMsg{Err: fmt.Errorf("grid server: %w", err)})
		c.mu.Lock()
		if c.state == Connected {
			c.state = Disconnected
			c.gridClient = nil
			c.gridServer = nil
			c.mu.Unlock()
			c.sendMsg(GridDisconnectedMsg{})
		} else {
			c.mu.Unlock()
		}
	}
}

// reconnectLoop periodically retries discovery when disconnected.
func (c *Controller) reconnectLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			state := c.state
			c.mu.Unlock()
			if state == Disconnected {
				debug.Printf("[grid] reconnectLoop: retrying discovery")
				c.discover()
			}
		}
	}
}

// buildSerialoscList constructs the raw OSC bytes for "/serialosc/list <host> <port>".
func buildSerialoscList(host string, port int) ([]byte, error) {
	msg := osc.NewMessage("/serialosc/list", host, int32(port))
	return msg.MarshalBinary()
}

// parseSerialoscDevice extracts device info from a /serialosc/device OSC message.
func parseSerialoscDevice(data []byte) (id, deviceType string, port int) {
	packet, err := osc.ParsePacket(string(data))
	if err != nil {
		return "", "", 0
	}
	m, ok := packet.(*osc.Message)
	if !ok {
		return "", "", 0
	}
	if m.Address != "/serialosc/device" || len(m.Arguments) < 3 {
		return "", "", 0
	}
	id, _ = m.Arguments[0].(string)
	deviceType, _ = m.Arguments[1].(string)
	p, _ := m.Arguments[2].(int32)
	return id, deviceType, int(p)
}
