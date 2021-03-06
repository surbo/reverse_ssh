package server

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/NHAS/reverse_ssh/internal"
	"golang.org/x/crypto/ssh"
)

func createSession(sshConn ssh.Conn, ptyReq, lastWindowChange ssh.Request) (sc ssh.Channel, err error) {

	splice, newrequests, err := sshConn.OpenChannel("session", nil)
	if err != nil {
		log.Printf("Unable to start remote session on host %s (%s) : %s\n", sshConn.RemoteAddr(), sshConn.ClientVersion(), err)
		return sc, fmt.Errorf("Unable to start remote session on host %s (%s) : %s", sshConn.RemoteAddr(), sshConn.ClientVersion(), err)
	}

	//Replay the pty and any the very last window size change in order to correctly size the PTY on the controlled client
	_, err = internal.SendRequest(ptyReq, splice)
	if err != nil {
		return sc, fmt.Errorf("Unable to send PTY request: %s", err)
	}

	_, err = internal.SendRequest(lastWindowChange, splice)
	if err != nil {
		return sc, fmt.Errorf("Unable to send last window change request: %s", err)
	}

	go ssh.DiscardRequests(newrequests)

	return splice, nil
}

func attachSession(newSession, currentClientSession ssh.Channel, currentClientRequests <-chan *ssh.Request) error {
	finished := make(chan bool)
	close := func() {
		newSession.Close()
		finished <- true // Stop the request passer on IO error
	}

	//Setup the pipes for stdin/stdout over the connections
	//newSession being the remote host being controlled
	var once sync.Once
	go func() {
		io.Copy(currentClientSession, newSession) // Potentially be more verbose about errors here
		once.Do(close)                            // Only close the newSession connection once

	}()
	go func() {
		io.Copy(newSession, currentClientSession)
		once.Do(close)
	}()
	defer once.Do(close)

RequestsPasser:
	for {
		select {
		case r := <-currentClientRequests:
			response, err := internal.SendRequest(*r, newSession)
			if err != nil {
				break RequestsPasser
			}

			if r.WantReply {
				r.Reply(response, nil)
			}
		case <-finished:
			break RequestsPasser
		}

	}

	return nil
}
