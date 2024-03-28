package rssh

import (
	"context"
	"errors"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
)

type SSHServer struct {
	options sshServerOptions
}

func NewSSHServer(options ...ServerOption) (*SSHServer, error) {

	serverOpts := sshServerOptions{}
	for _, option := range options {
		if err := option(&serverOpts); err != nil {
			return nil, err
		}
	}

	if err := serverOpts.check(); err != nil {
		return nil, err
	}

	return &SSHServer{options: serverOpts}, nil
}

func (s *SSHServer) Listen(ctx context.Context) error {
	listener, err := net.ListenTCP("tcp", s.options.address)
	if err != nil {
		return err
	}

	config := ssh.ServerConfig{
		PublicKeyCallback: s.options.publicKeyCallback,
	}
	for _, key := range s.options.hostKeys {
		config.AddHostKey(key)
	}

	ch := make(chan net.Conn)
	defer close(ch)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}

				log.Println(err)
				continue
			}

			ch <- conn
		}
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case conn := <-ch:
			go func() {
				sshConn, newChans, reqsChan, err := ssh.NewServerConn(conn, &config)
				if err != nil {
					log.Fatalf("fail to establish ssh connection: %v\n", err)
				}
				defer func(sshConn *ssh.ServerConn) {
					err := sshConn.Close()
					if err != nil {
						log.Printf("fail to close ssh connection: %v\n", err)
					}
				}(sshConn)
				defer func(conn net.Conn) {
					err := conn.Close()
					if err != nil {
						log.Printf("fail to close connection: %v\n", err)
					}
				}(conn)

				go s.options.requestHandler(reqsChan)

				for chanReq := range newChans {
					go s.options.newChannelHandler(ctx, sshConn, chanReq)
				}
			}()
		}
	}

	if ctx.Err() != nil && !errors.Is(ctx.Err(), context.Canceled) {
		return ctx.Err()
	}

	return nil
}

type SSHServerError string

func (e SSHServerError) Error() string {
	return string(e)
}

func (e SSHServerError) String() string {
	return string(e)
}

type SSHIdentifier struct {
	Identifier string
}

type SSHIdentifierResolver interface {
	Resolve(identifier string) (*net.TCPAddr, error)
}
