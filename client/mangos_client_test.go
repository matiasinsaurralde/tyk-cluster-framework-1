package client

import (
	"github.com/TykTechnologies/tyk-cluster-framework/encoding"
	"github.com/TykTechnologies/tyk-cluster-framework/payloads"
	"github.com/TykTechnologies/tyk-cluster-framework/server"
	"testing"
	"time"
)

type testPayloadData struct {
	FullName string
}

func TestMangosClient(t *testing.T) {
	var s, s1 server.Server
	var err error
	if s, err = server.NewServer("mangos://127.0.0.1:9100", encoding.JSON); err != nil {
		t.Fatal(err)
	}

	if s1, err = server.NewServer("mangos://127.0.0.1:9105", encoding.JSON); err != nil {
		t.Fatal(err)
	}

	s.Listen()
	s1.Listen()

	time.Sleep(1 * time.Second)

	// Test pub/sub
	t.Run("Client Side Publish", func(t *testing.T) {
		var err error
		var c Client

		ch := "tcf.test.mangos-server.client-publish"
		msg := "Tyk Cluster Framework: Client"
		resultChan := make(chan testPayloadData)

		if c, err = NewClient("mangos://127.0.0.1:9100", encoding.JSON); err != nil {
			t.Fatal(err)
		}

		// Connect
		if err = c.Connect(); err != nil {
			t.Fatal(err)
		}

		// Subscribe to some stuff
		var subChan chan string
		if subChan, err = c.Subscribe(ch, func(payload payloads.Payload) {
			var d testPayloadData
			err := payload.DecodeMessage(&d)
			if err != nil {
				t.Fatalf("Decode payload failed: %v", err)
			}

			resultChan <- d
		}); err != nil {
			t.Fatal(err)
		}

		var dp payloads.Payload
		if dp, err = payloads.NewPayload(testPayloadData{msg}); err != nil {
			t.Fatal(err)
		}

		select {
		case s := <-subChan:
			if s != ch {
				t.Fatal("Incorrect subscribe channel returned!")
			}
		case <-time.After(time.Millisecond * 500):
			t.Fatalf("Channel wait timed out")
		}

		time.Sleep(1 * time.Second)

		// This is ugly, but mangos handles connect in the background, so we need to wait :-/
		if err = c.Publish(ch, dp); err != nil {
			t.Fatal(err)
		}

		select {
		case v := <-resultChan:
			if v.FullName != msg {
				t.Fatalf("Unexpected return value: %v", v)
			}
		case <-time.After(time.Second * 3):
			t.Fatalf("Timed out")
		}

		c.Stop()
		time.Sleep(1 * time.Second)
		close(resultChan)
	})

	// Test multiple subs with a single client
	t.Run("Multiple Client Side Subs", func(t *testing.T) {
		var err error
		var c1 Client

		ch1 := "tcf.test.mangos-server.client-sub1"
		ch2 := "tcf.test.mangos-server.client-sub2"

		resultChan1 := make(chan testPayloadData)
		resultChan2 := make(chan testPayloadData)

		if c1, err = NewClient("mangos://127.0.0.1:9100", encoding.JSON); err != nil {
			t.Fatal(err)
		}

		// Connect
		if err = c1.Connect(); err != nil {
			t.Fatal(err)
		}

		// Subscribe to some stuff
		var subChan chan string
		if subChan, err = c1.Subscribe(ch1, func(payload payloads.Payload) {
			var d testPayloadData
			err := payload.DecodeMessage(&d)
			if err != nil {
				t.Fatalf("Decode payload failed: %v", err)
			}

			resultChan1 <- d
		}); err != nil {
			t.Fatal("err")
		}

		// Subscribe to some stuff
		if subChan, err = c1.Subscribe(ch2, func(payload payloads.Payload) {
			var d testPayloadData
			err := payload.DecodeMessage(&d)
			if err != nil {
				t.Fatalf("Decode payload failed: %v", err)
			}

			resultChan2 <- d
		}); err != nil {
			t.Fatal("err")
		}

		var dpChan1, dpChan2 payloads.Payload
		ch1Msg := "Channel 1"
		ch2Msg := "Channel 2"
		if dpChan1, err = payloads.NewPayload(testPayloadData{ch1Msg}); err != nil {
			t.Fatal(err)
		}
		if dpChan2, err = payloads.NewPayload(testPayloadData{ch2Msg}); err != nil {
			t.Fatal(err)
		}

		select {
		case s := <-subChan:
			if s != ch1 && s != ch2 {
				t.Fatalf("Incorrect subscribe channel returned: %v", s)
			}
		case <-time.After(time.Millisecond * 100):
			t.Fatalf("Channel wait timed out")
		}

		select {
		case s := <-subChan:
			if s != ch1 && s != ch2 {
				t.Fatalf("Incorrect subscribe channel returned: %v", s)
			}
		case <-time.After(time.Millisecond * 100):
			t.Fatalf("Channel wait timed out")
		}

		// This is ugly, but mangos handles connect in the background, so we need to wait :-/
		time.Sleep(1 * time.Second)
		if err = c1.Publish(ch1, dpChan1); err != nil {
			t.Fatal(err)
		}
		if err = c1.Publish(ch2, dpChan2); err != nil {
			t.Fatal(err)
		}

		// Inverted result channels here so we can test async
		select {
		case v := <-resultChan2:
			if v.FullName != ch2Msg {
				t.Fatalf("Unexpected return value: %v", v)
			}
		case <-time.After(time.Millisecond * 500):
			t.Fatalf("Chan 2 timed out")
		}

		select {
		case v := <-resultChan1:
			if v.FullName != ch1Msg {
				t.Fatalf("Unexpected return value: %v", v)
			}
		case <-time.After(time.Millisecond * 500):
			t.Fatalf("Chan 1 timed out")
		}

		c1.Stop()

		time.Sleep(1 * time.Second)
	})

	//// Test multiple subscribes with one client, but ignore the subs notification channel
	//// because it might be unused for brevity
	t.Run("Multiple Client Side Subs, but ignore sub channel", func(t *testing.T) {
		var err error
		var c1 Client

		ch1 := "tcf.test.mangos-server.client-sub1"
		ch2 := "tcf.test.mangos-server.client-sub2"

		resultChan3 := make(chan testPayloadData)
		resultChan4 := make(chan testPayloadData)

		if c1, err = NewClient("mangos://127.0.0.1:9100", encoding.JSON); err != nil {
			t.Fatal(err)
		}

		// Connect
		if err = c1.Connect(); err != nil {
			t.Fatal(err)
		}

		// Subscribe to some stuff
		if _, err = c1.Subscribe(ch1, func(payload payloads.Payload) {
			var d testPayloadData
			err := payload.DecodeMessage(&d)
			if err != nil {
				t.Fatalf("Decode payload failed: %v", err)
			}

			resultChan3 <- d
		}); err != nil {
			t.Fatal("err")
		}

		// Subscribe to some stuff
		if _, err = c1.Subscribe(ch2, func(payload payloads.Payload) {
			var d testPayloadData
			err := payload.DecodeMessage(&d)
			if err != nil {
				t.Fatalf("Decode payload failed: %v", err)
			}

			resultChan4 <- d
		}); err != nil {
			t.Fatal("err")
		}

		var dpChan1, dpChan2 payloads.Payload
		ch1Msg := "Channel 1"
		ch2Msg := "Channel 2"
		if dpChan1, err = payloads.NewPayload(testPayloadData{ch1Msg}); err != nil {
			t.Fatal(err)
		}
		if dpChan2, err = payloads.NewPayload(testPayloadData{ch2Msg}); err != nil {
			t.Fatal(err)
		}

		// This is ugly, but mangos handles connect in the background, so we need to wait :-/
		time.Sleep(1 * time.Second)

		if err = c1.Publish(ch1, dpChan1); err != nil {
			t.Fatal(err)
		}
		if err = c1.Publish(ch2, dpChan2); err != nil {
			t.Fatal(err)
		}

		// Inverted result channels here so we can test async
		select {
		case v := <-resultChan4:
			if v.FullName != ch2Msg {
				t.Fatalf("Unexpected return value: %v", v)
			}
		case <-time.After(time.Millisecond * 500):
			t.Fatalf("Chan 2 timed out")
		}

		select {
		case v := <-resultChan3:
			if v.FullName != ch1Msg {
				t.Fatalf("Unexpected return value: %v", v)
			}
		case <-time.After(time.Millisecond * 500):
			t.Fatalf("Chan 1 timed out")
		}

		c1.Stop()
	})

	t.Run("Broadcast Test", func(t *testing.T) {
		var b Client
		var err error
		resultChan := make(chan testPayloadData)

		if b, err = NewClient("mangos://127.0.0.1:9105", encoding.JSON); err != nil {
			t.Fatal(err)
		}

		// Connect
		if err = b.Connect(); err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second)

		ch := "tcf.test.mangos-server.broadcast-test"
		chMsg := "Channel 1"
		var pl payloads.Payload
		if pl, err = payloads.NewPayload(testPayloadData{chMsg}); err != nil {
			t.Fatal(err)
		}

		if err := b.Broadcast(ch, pl, 1); err != nil {
			t.Fatal(err)
		}

		if _, err = b.Subscribe(ch, func(payload payloads.Payload) {
			var d testPayloadData
			err := payload.DecodeMessage(&d)
			if err != nil {
				t.Fatalf("Decode payload failed: %v", err)
			}

			resultChan <- d
		}); err != nil {
			t.Fatal(err)
		}

		loopCnt := 0
		msgCnt := 0
		for loopCnt = 0; loopCnt <= 3; loopCnt++ {
			select {
			case v := <-resultChan:
				if v.FullName != chMsg {
					t.Fatalf("Unexpected return value: %v", v)
				}
				msgCnt += 1
			case <-time.After(time.Second * 2):
				// fall through
			}
		}

		if msgCnt == 0 {
			t.Fatal("Received no messages")
		}

		if msgCnt < 3 {
			t.Fatalf("Received less than 3 messages, got: %v", msgCnt)
		}

		b.Stop()
	})

	t.Run("Broadcast Test Change Payload", func(t *testing.T) {
		var b Client
		var err error
		resultChan := make(chan testPayloadData)

		if b, err = NewClient("mangos://127.0.0.1:9105", encoding.JSON); err != nil {
			t.Fatal(err)
		}

		// Connect
		if err = b.Connect(); err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second)

		ch := "tcf.test.mangos-server.broadcast-test"
		chMsg := "Channel 1"
		var pl payloads.Payload
		if pl, err = payloads.NewPayload(testPayloadData{chMsg}); err != nil {
			t.Fatal(err)
		}

		if err := b.Broadcast(ch, pl, 1); err != nil {
			t.Fatal(err)
		}

		if _, err = b.Subscribe(ch, func(payload payloads.Payload) {
			var d testPayloadData
			err := payload.DecodeMessage(&d)
			if err != nil {
				t.Fatalf("Decode payload failed: %v", err)
			}

			resultChan <- d
		}); err != nil {
			t.Fatal(err)
		}

		loopCnt := 0
		msgCnt := 0
		for loopCnt = 0; loopCnt <= 3; loopCnt++ {
			select {
			case v := <-resultChan:
				if v.FullName != chMsg {
					t.Fatalf("Unexpected return value: %v", v)
				}
				msgCnt += 1
			case <-time.After(time.Second * 2):
				// fall through
			}
		}

		if msgCnt == 0 {
			t.Fatal("Received no messages")
		}

		if msgCnt < 3 {
			t.Fatalf("Received less than 3 messages, got: %v", msgCnt)
		}

		b.Stop()
	})

	// Test stop
	time.Sleep(1 * time.Second)
	if err = s.Stop(); err != nil {
		t.Fatal(err)
	}

}
