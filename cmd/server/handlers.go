package main

import (
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"ssh-reverse-proxy/rssh"
)

func newChannelHandler(ctx context.Context, sshConn *ssh.ServerConn, newChannel ssh.NewChannel) {
	switch newChannel.ChannelType() {
	case "session":
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go handleSession(ctx, sshConn, channel, requests, newChannel.ExtraData())
	case "direct-tcpip":
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go handleDirectTCPIP(channel, requests, newChannel.ExtraData())

	default:
		err := newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", newChannel.ChannelType()))
		if err != nil {
			log.Println(err)
			return
		}
		return
	}
}

func handleSession(ctx context.Context, clientConn *ssh.Client, channel ssh.Channel, requests <-chan *ssh.Request, extraData []byte) {
	defer func(channel ssh.Channel) {
		err := channel.Close()
		if err != nil {
			log.Println(err)
		}
	}(channel)

	clientSession, err := clientConn.NewSession()
	if err != nil {
		log.Println(err)
		return
	}

	for req := range requests {
		switch req.Type {
		case "shell":
			go handleSessionShell(ctx, clientSession, req)
		case "pty-req":
			go handleSessionPtyReq(req)
		case "env":
			go handleSessionEnv(req)
		case "window-change":
			go handleSessionWindowChange(clientConn, req)
		default:
			_, err := channel.Write([]byte("unsupported request type\n"))
			if err != nil {
				log.Println(err)
			}
			return
		}
		if req.WantReply {
			err := req.Reply(true, nil)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func requestsHandler(in <-chan *ssh.Request) {
	for req := range in {
		log.Printf("request type: %s\n", req.Type)
		err := req.Reply(false, nil)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func handleDirectTCPIP(channel ssh.Channel, requests <-chan *ssh.Request, extraData []byte) {
	defer func(channel ssh.Channel) {
		err := channel.Close()
		if err != nil {
			log.Println(err)
		}
	}(channel)

	var err error

	directTcpIp, err := rssh.ParseDirectTCPIP(extraData)
	if err != nil {
		if err != nil {
			errStr := fmt.Sprintf("error setting window size: %s\n", err)
			log.Println(errStr)
			_, err = channel.Write([]byte(errStr))
			if err != nil {
				log.Println(err)
			}
			return
		}
	}

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", directTcpIp.HostToConnect, directTcpIp.PortToConnect))
	if err != nil {
		errStr := fmt.Sprintf("error resolving tcp address: %s\n", err)
		log.Println(errStr)
		_, err = channel.Write([]byte(errStr))
		if err != nil {
			log.Println(err)
		}
		return
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		errStr := fmt.Sprintf("error dialing tcp address: %s\n", err)
		log.Println(errStr)
		_, err = channel.Write([]byte(errStr))
		if err != nil {
			log.Println(err)
		}
		return
	}

	go func() {
		_, err := io.Copy(channel, conn)
		if err != nil {
			log.Println(err)
			return
		}
	}()
	go func() {
		_, err := io.Copy(conn, channel)
		if err != nil {
			log.Println(err)
			return
		}
	}()

	for req := range requests {
		log.Printf("request type: %s\n", req.Type)
	}
}

func handleSessionShell(ctx context.Context, clientSession *ssh.Session, req *ssh.Request) {
	err := clientSession.Shell()
	if err != nil {
		log.Println(err)
		return
	}

	err = clientSession.Wait()
}

func handleSessionExec(ctx context.Context, clientSession *ssh.Session, req *ssh.Request) {
	command, _, err := rssh.ParseNextString(req.Payload, 0)
	if err != nil {
		log.Println(err)
		return
	}

	err = clientSession.Run(command)
	if err != nil {
		log.Println(err)
		return
	}

	err = clientSession.Wait()
}

func handleSessionEnv(clientConn *ssh.Client, req *ssh.Request) {

}

func handleSessionPtyReq(clientConn *ssh.Client, req *ssh.Request) {

}

func handleSessionWindowChange(ctx context.Context, clientSession *ssh.Session, req *ssh.Request) {
	//columns := rssh.ArrayToUint32([4]byte(req.Payload[0:4]))
	//rows := rssh.ArrayToUint32([4]byte(req.Payload[4:8]))
	width := rssh.ArrayToUint32([4]byte(req.Payload[8:12]))
	height := rssh.ArrayToUint32([4]byte(req.Payload[12:16]))

	err := clientSession.WindowChange(int(height), int(width))
	if err != nil {
		log.Println(err)
		return
	}
}

func publicKeyCallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	log.Printf("public key for %s\n", conn.User())

	return nil, nil
}
