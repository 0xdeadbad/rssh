package main

import (
	"context"
	"errors"
	"github.com/pkg/profile"
	"log"
	"os"
	"os/signal"
	"ssh-reverse-proxy/rssh"
)

func main() {
	p := profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	defer p.Stop()

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	sshServer, err := rssh.NewSSHServer(rssh.WithPrivateKeyPath("./id_rsa"), rssh.WithTCPAddress("0.0.0.0:51022"), rssh.WithPublicKeyCallback(publicKeyCallback), rssh.WithNewChannelHandler(newChannelHandler), rssh.WithRequestsHandler(requestsHandler))
	if err != nil {
		log.Fatalln(err)
	}

	if err := sshServer.Listen(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalln(err)
	}
}
