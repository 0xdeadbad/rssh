package rssh

import (
	"context"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
)

type sshServerOptions struct {
	address           *net.TCPAddr
	hostKeys          []ssh.Signer
	requestHandler    SSHRequestHandler
	newChannelHandler SSHNewChannelHandler
	publicKeyCallback SSHPublicKeyCallback
	resolver          SSHIdentifierResolver
}

func (o *sshServerOptions) check() error {
	if o.address == nil {
		address, err := net.ResolveTCPAddr("tcp", "0.0.0.0:22")
		if err != nil {
			return err
		}
		o.address = address
	}
	if len(o.hostKeys) == 0 {
		rsaKey, err := generatePrivateKey(2048)
		if err != nil {
			return err
		}
		rsaSigner, err := ssh.NewSignerFromKey(rsaKey)
		if err != nil {
			return err
		}
		o.hostKeys = append(o.hostKeys, rsaSigner)
	}
	if o.requestHandler == nil {
		o.requestHandler = ssh.DiscardRequests
	}
	if o.publicKeyCallback == nil {
		return ErrPublicKeyCallbackNotSet
	}
	if o.newChannelHandler == nil {
		return ErrNewChannelHandlerNotSet
	}
	return nil
}

const (
	ErrPublicKeyCallbackNotSet SSHServerError = "public key callback is not set"
	ErrNewChannelHandlerNotSet SSHServerError = "new channel handler is not set"
	ErrSSHConnectionFail       SSHServerError = "failed to establish ssh connection"
)

type SSHRequestHandler func(in <-chan *ssh.Request)
type SSHPublicKeyCallback func(conn ssh.ConnMetadata, publicKey ssh.PublicKey) (*ssh.Permissions, error)
type SSHNewChannelHandler func(ctx context.Context, sshConn *ssh.ServerConn, newChannel ssh.NewChannel)

type ServerOption func(*sshServerOptions) error

func WithTCPAddress(address string) ServerOption {
	return func(o *sshServerOptions) error {
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			return err
		}

		o.address = addr

		return nil
	}
}

func WithResolver(resolver SSHIdentifierResolver) ServerOption {
	return func(o *sshServerOptions) error {
		o.resolver = resolver

		return nil
	}
}

func WithPrivateKeyPath(keyPath string) ServerOption {
	return func(o *sshServerOptions) error {
		privBytes, err := os.ReadFile(keyPath)
		if err != nil {
			return err
		}

		privKey, err := ssh.ParsePrivateKey(privBytes)
		if err != nil {
			return err
		}

		o.hostKeys = append(o.hostKeys, privKey)

		return nil
	}
}

func WithPrivateKey(privKey ssh.Signer) ServerOption {
	return func(o *sshServerOptions) error {
		o.hostKeys = append(o.hostKeys, privKey)

		return nil
	}
}

func WithRequestsHandler(requestHandler SSHRequestHandler) ServerOption {
	return func(o *sshServerOptions) error {
		o.requestHandler = requestHandler

		return nil
	}
}

func WithNewChannelHandler(newChannel SSHNewChannelHandler) ServerOption {
	return func(o *sshServerOptions) error {
		o.newChannelHandler = newChannel

		return nil
	}
}

func WithPublicKeyCallback(publicKeyCallback SSHPublicKeyCallback) ServerOption {
	return func(o *sshServerOptions) error {
		o.publicKeyCallback = publicKeyCallback

		return nil
	}
}
