package docker

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"github.com/dotcloud/docker"
	"github.com/dotcloud/docker/engine"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"testing"
	"time"
)

const testDaemonHttpsAddr = "localhost:4271"

// TestHttpsInfo connects via two-way authenticated HTTPS to the info endpoint
func TestHttpsInfo(t *testing.T) {
	root, err := newTestDirectory(unitTestStoreBase)
	if err != nil {
		log.Fatal(err)
	}
	eng, err := engine.New(root)
	if err != nil {
		log.Fatal(err)
	}
	job := eng.Job("initapi")
	job.Setenv("Root", root)
	job.SetenvBool("Tls", true)
	job.SetenvBool("TlsVerify", true)
	job.Setenv("TlsCa", "fixtures/https/ca.crt")
	job.Setenv("TlsCert", "fixtures/https/server.crt")
	job.Setenv("TlsKey", "fixtures/https/server.key")
	if err := job.Run(); err != nil {
		log.Fatal(err)
	}

	listenURL := &url.URL{
		Scheme: testDaemonProto,
		Host:   testDaemonHttpsAddr,
	}
	job = eng.Job("serveapi", listenURL.String())

	go func() {
		if err := job.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(500 * time.Millisecond)

	stdout, stdoutPipe := io.Pipe()

	certPool := x509.NewCertPool()
	file, err := ioutil.ReadFile(testTlsCa)
	if err != nil {
		t.Fatal(err)
	}
	certPool.AppendCertsFromPEM(file)

	cert, err := tls.LoadX509KeyPair("fixtures/https/client.crt", "fixtures/https/client.key")
	if err != nil {
		t.Fatalf("Couldn't load X509 key pair: %s", err)
	}
	tlsConfig := &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{cert},
	}

	cli := docker.NewDockerCli(nil, stdoutPipe, ioutil.Discard, testDaemonProto, testDaemonHttpsAddr, tlsConfig)

	// defer cleanup(globalEngine, t)
	c := make(chan struct{})
	go func() {
		defer close(c)
		if err := cli.CmdInfo(); err != nil {
			t.Fatal(err)
		}
	}()

	setTimeout(t, "Reading command output time out", 2*time.Second, func() {
		if _, err := bufio.NewReader(stdout).ReadString('\n'); err != nil {
			t.Fatal(err)
		}
	})

	// Cleanup pipes
	if err := closeWrap(stdout, stdoutPipe); err != nil {
		t.Fatal(err)
	}
}
