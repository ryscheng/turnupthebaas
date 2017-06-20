package libtalek

import "encoding/binary"

// message represents a emitted message by talek. It may be split into
// multiple parts by the client for transmission, and reassembled before
// being sent to the application.
type message struct {
	contents []byte

	// TODO: there's a cute data structure for efficiently tracking out of order
	// receipt of messages. This is not that.
	receivedEnd uint32
}

// fragmentHeaderLength encodes the length of a message fragment header
const fragmentHeaderLength = 5

// fragmentHeader encodes the header of each wire fragment a message is split into.
type fragmentHeader struct {
	flag byte
	left uint32
}

// fromBytes reconstructs a fragmentHeader from its wire format
func fromBytes(message []byte) *fragmentHeader {
	if len(message) < fragmentHeaderLength {
		return nil
	}
	f := new(fragmentHeader)
	f.flag = message[0]
	f.left = binary.LittleEndian.Uint32(message[1:5])
	return f
}

// IsNewMessage indicates if this fragment represents the first fragment in a message
func (f *fragmentHeader) IsNewMessage() bool {
	return (f.flag & 1) == 1
}

func newFragment(firstFragment bool, remainingLength uint32) *fragmentHeader {
	f := new(fragmentHeader)
	f.left = remainingLength
	if firstFragment {
		f.flag |= 1
	}
	return f
}

// ToBytes serializes a fragment header to a byte slice
func (f *fragmentHeader) ToBytes(buf []byte) {
	buf[0] = f.flag
	binary.LittleEndian.PutUint32(buf[1:5], f.left)
}

// newMessage creates a message from an underlying byte slice
func newMessage(msg []byte) *message {
	m := new(message)
	m.contents = msg
	m.receivedEnd = uint32(len(m.contents))
	return m
}

// Split divides a full message into a set of parts no larger than partSize.
func (m *message) Split(partSize int) [][]byte {
	denom := (partSize - fragmentHeaderLength)
	numMsgs := len(m.contents) / denom
	if len(m.contents)%denom != 0 {
		numMsgs++
	}
	messages := make([][]byte, numMsgs)

	contentLength := len(m.contents)
	remaining := len(m.contents)
	for i := 0; i < len(messages); i++ {
		part := make([]byte, partSize)
		header := newFragment(i == 0, uint32(remaining))
		header.ToBytes(part)
		remaining -= copy(part[fragmentHeaderLength:], m.contents[contentLength-remaining:])

		messages[i] = part
	}

	return messages
}

// Join Adds a newly received part to a partially reconstructed message
func (m *message) Join(part []byte) bool {
	header := fromBytes(part)
	if header == nil {
		return false
	}
	if header.IsNewMessage() && m.receivedEnd == 0 {
		m.contents = make([]byte, header.left)
	}
	if len(m.contents)-int(m.receivedEnd) != int(header.left) {
		return false
	}
	copy(m.contents[m.receivedEnd:], part[fragmentHeaderLength:])
	m.receivedEnd += uint32(len(part)) - fragmentHeaderLength
	return m.receivedEnd >= uint32(len(m.contents))
}

// Retrieve provides the underlying bytes of a message when known.
func (m *message) Retrieve() []byte {
	if m.receivedEnd >= uint32(len(m.contents)) {
		return m.contents
	}
	return nil
}
