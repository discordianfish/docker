package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/dotcloud/docker"
	"github.com/dotcloud/docker/engine"
	"github.com/dotcloud/docker/sysinit"
	"github.com/dotcloud/docker/utils"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	GITCOMMIT string
	VERSION   string
)

func main() {
	if selfPath := utils.SelfPath(); selfPath == "/sbin/init" || selfPath == "/.dockerinit" {
		// Running in init mode
		sysinit.SysInit()
		return
	}

	var (
		flVersion            = flag.Bool("v", false, "Print version information and quit")
		flDaemon             = flag.Bool("d", false, "Enable daemon mode")
		flDebug              = flag.Bool("D", false, "Enable debug mode")
		flAutoRestart        = flag.Bool("r", true, "Restart previously running containers")
		bridgeName           = flag.String("b", "", "Attach containers to a pre-existing network bridge; use 'none' to disable container networking")
		bridgeIp             = flag.String("bip", "", "Use this CIDR notation address for the network bridge's IP, not compatible with -b")
		pidfile              = flag.String("p", "/var/run/docker.pid", "Path to use for daemon PID file")
		flRoot               = flag.String("g", "/var/lib/docker", "Path to use as the root of the docker runtime")
		flEnableCors         = flag.Bool("api-enable-cors", false, "Enable CORS headers in the remote API")
		flDns                = docker.NewListOpts(docker.ValidateIp4Address)
		flEnableIptables     = flag.Bool("iptables", true, "Disable docker's addition of iptables rules")
		flDefaultIp          = flag.String("ip", "0.0.0.0", "Default IP address to use when binding container ports")
		flInterContainerComm = flag.Bool("icc", true, "Enable inter-container communication")
		flGraphDriver        = flag.String("s", "", "Force the docker runtime to use a specific storage driver")
		flHosts              = docker.NewListOpts(docker.ValidateHost)
		flMtu                = flag.Int("mtu", docker.DefaultNetworkMtu, "Set the containers network mtu")
		flTls                = flag.Bool("tls", false, "Use TLS; implied by all other tls* flags")
		flCa                 = flag.String("tlscacert", "", "Trust only remotes providing a certificate signed by the CA given here")
		flCert               = flag.String("tlscert", "", "Path to TLS certificate file")
		flKey                = flag.String("tlskey", "", "Path to TLS key file")
	)
	flag.Var(&flDns, "dns", "Force docker to use specific DNS servers")
	flag.Var(&flHosts, "H", "Multiple tcp://host:port or unix://path/to/socket to bind in daemon mode, single connection otherwise")

	flag.Parse()

	if *flVersion {
		showVersion()
		return
	}
	if flHosts.Len() == 0 {
		defaultHost := os.Getenv("DOCKER_HOST")

		if defaultHost == "" || *flDaemon {
			// If we do not have a host, default to unix socket
			defaultHost = fmt.Sprintf("unix://%s", docker.DEFAULTUNIXSOCKET)
		}
		flHosts.Set(defaultHost)
	}

	if *bridgeName != "" && *bridgeIp != "" {
		log.Fatal("You specified -b & -bip, mutually exclusive options. Please specify only one.")
	}

	if (len(*flKey) > 0 && len(*flCert) == 0) || (len(*flKey) == 0 && len(*flCert) > 0) {
		log.Fatalf("Need TLS cert and key to enable TLS authentication")
	}

	if *flDebug {
		os.Setenv("DEBUG", "1")
	}

	docker.GITCOMMIT = GITCOMMIT
	docker.VERSION = VERSION
	if *flDaemon {
		if flag.NArg() != 0 {
			flag.Usage()
			return
		}

		eng, err := engine.New(*flRoot)
		if err != nil {
			log.Fatal(err)
		}
		// Load plugin: httpapi
		job := eng.Job("initapi")
		job.Setenv("Pidfile", *pidfile)
		job.Setenv("Root", *flRoot)
		job.SetenvBool("AutoRestart", *flAutoRestart)
		job.SetenvBool("EnableCors", *flEnableCors)
		job.SetenvList("Dns", flDns.GetAll())
		job.SetenvBool("EnableIptables", *flEnableIptables)
		job.Setenv("BridgeIface", *bridgeName)
		job.Setenv("BridgeIp", *bridgeIp)
		job.Setenv("DefaultIp", *flDefaultIp)
		job.SetenvBool("InterContainerCommunication", *flInterContainerComm)
		job.Setenv("GraphDriver", *flGraphDriver)
		job.SetenvInt("Mtu", *flMtu)
		job.Setenv("TlsCa", *flCa)
		job.Setenv("TlsCert", *flCert)
		job.Setenv("TlsKey", *flKey)
		if err := job.Run(); err != nil {
			log.Fatal(err)
		}
		// Serve api
		job = eng.Job("serveapi", flHosts.GetAll()...)
		job.SetenvBool("Logging", true)
		if err := job.Run(); err != nil {
			log.Fatal(err)
		}
	} else {
		if flHosts.Len() > 1 {
			log.Fatal("Please specify only one -H")
		}
		protoAddrParts := strings.SplitN(flHosts.GetAll()[0], "://", 2)

		var (
			errc error
			tlsConfig tls.Config
		)

	  if len(*flCert) > 0 && len(*flKey) > 0 {
			*flTls = true
			cert, err := tls.LoadX509KeyPair(*flCert, *flKey)
			if err != nil {
				log.Fatalf("Couldn't load X509 key pair: %s", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		if len(*flCa) > 0 {
			*flTls = true
			certPool := x509.NewCertPool()
			file, err := ioutil.ReadFile(*flCa)
			if err != nil {
				log.Fatal(err)
			}
			certPool.AppendCertsFromPEM(file)
			tlsConfig.RootCAs = certPool
		}

		if *flTls {
				errc = docker.ParseCommands(protoAddrParts[0], protoAddrParts[1], &tlsConfig, flag.Args()...)
			} else {
				errc = docker.ParseCommands(protoAddrParts[0], protoAddrParts[1], nil, flag.Args()...)
		}
		if errc != nil {
			if sterr, ok := errc.(*utils.StatusError); ok {
				if sterr.Status != "" {
					log.Println(sterr.Status)
				}
				os.Exit(sterr.StatusCode)
			}
			log.Fatal(errc)
		}
	}
}

func showVersion() {
	fmt.Printf("Docker version %s, build %s\n", VERSION, GITCOMMIT)
}
