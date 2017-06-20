package libtalek

import "testing"

func TestMessage(t *testing.T) {
	theMsg := make([]byte, 1024*256)

	msg := newMessage(theMsg)

	// Should be a full msg
	retr := msg.Retrieve()
	if retr == nil || len(retr) != len(theMsg) {
		t.Fatalf("failed to retrieve message")
	}

	// Split
	parts := msg.Split(256)

	if len(parts) < 1024 {
		t.Fatalf("Failed to split")
	}

	// reconstruct
	recon := message{}

	for i := 0; i < len(parts); i++ {
		ret := recon.Join(parts[i])
		if i < len(parts)-1 && ret == true {
			t.Fatalf("indicated message reconstructed too early")
		}
		if i == len(parts)-1 && ret != true {
			t.Fatalf("didn't indicate message reconstructed")
		}
	}
	reconmsg := recon.Retrieve()
	if reconmsg == nil || len(reconmsg) != len(theMsg) {
		t.Fatalf("failed to reconstruct split msg")
	}
}
