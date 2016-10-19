package pir

import "fmt"
import "math/rand"
import "testing"

func getSocket() string {
  return fmt.Sprintf("pirtest%d.socket", rand.Int())
}

func TestConnnect(t *testing.T) {
  sockName := getSocket()
  status := make(chan int)
	go CreateMockServer(status, sockName)
  <- status

	_, err := Connect(sockName)
	if err != nil {
		t.Error(err)
    return
	}

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
  db.DB[0] = 1

  pirServer.SetDB(db)

  masks := make([]byte, 512)
  masks[0] = 1

  response, err := pirServer.Read(masks)

  if err != nil || response == nil || response[0] !=1 {
    t.Error(err)
  }

  status <- 1
}
