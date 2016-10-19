package pir

import "fmt"
import "math/rand"
import "os"
import "testing"

func getSocket() string {
  if os.Getenv("PIR_SOCKET") != "" {
    fmt.Printf("Testing against running pird at %s.\n", os.Getenv("PIR_SOCKET"))
    return os.Getenv("PIR_SOCKET")
  }
  return fmt.Sprintf("pirtest%d.socket", rand.Int())
}

func TestConnnect(t *testing.T) {
  sockName := getSocket()
  status := make(chan int)
	go CreateMockServer(status, sockName)
  <- status

	pirServer, err := Connect(sockName)
	if err != nil {
		t.Error(err)
    return
	}

  pirServer.Disconnect()

  status <- 1
}

func TestPir(t *testing.T) {
  sockName := getSocket()
  status := make(chan int)
	go CreateMockServer(status, sockName)
  <- status

	pirServer, err := Connect(sockName)
	if err != nil {
		t.Error(err)
    return
	}

  pirServer.Configure(512, 512, 8)
  db, err := pirServer.GetDB()
  if err != nil {
    t.Error(err)
    return
  }
  for x := range db.DB {
    db.DB[x] = byte(x)
  }

  pirServer.SetDB(db)

  masks := make([]byte, 512)
  masks[0] = 0x01

  response, err := pirServer.Read(masks)

  if err != nil || response == nil {
    t.Error(err)
    return
  }

  if response[1] != byte(1) {
    t.Error(fmt.Sprintf("Response is incorrect. byte 1 was %d, not '1'.", response[1]))
  }

  pirServer.Disconnect()

  status <- 1
}

func BenchmarkPir(b *testing.B) {
  cellLength := 1024
  cellCount := 2048
  batchSize := 8

  sockName := getSocket()
  status := make(chan int)
	go CreateMockServer(status, sockName)
  <- status

	pirServer, err := Connect(sockName)
	if err != nil {
		b.Error(err)
    return
	}

  pirServer.Configure(cellLength, cellCount, batchSize)
  db, err := pirServer.GetDB()
  if err != nil {
    b.Error(err)
    return
  }
  for x := range db.DB {
    db.DB[x] = byte(x)
  }

  pirServer.SetDB(db)

  masks := make([]byte, cellCount * batchSize / 8)
  masks[0] = 0x01

  b.ResetTimer()

  for i := 0; i < b.N; i++ {
    response, err := pirServer.Read(masks)

    if err != nil || response == nil {
      b.Error(err)
    }
    b.SetBytes(int64(len(response)))
  }

  pirServer.Disconnect()

  status <- 1
}
