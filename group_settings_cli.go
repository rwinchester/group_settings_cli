package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"github.com/op/go-logging"
	"os"

	// "github.com/digitallumens/turnpike"
	// "github.com/manyminds/api2go/jsonapi"
	// "github.com/mitchellh/mapstructure"
	"github.com/urfave/cli"

	backend "github.com/digitallumens/lightworks_backend/client"
	// models "github.com/digitallumens/lightworks_backend/models"
	router_client "github.com/digitallumens/lightworks_router/client"
)

type GroupSettingsCommand struct {
	Token     string                 `json:"token" structs:"token" mapstructure:"token"`
	GroupID   string                 `json:"id" structs:"id" mapstructure:"id"`
	Procedure string                 `json:"procedure" structs:"procedure" mapstructure:"procedure"`
	Params    map[string]interface{} `json:"params" structs:"params" mapstructure:"params"`
}

type GroupSettingsCommandReply struct {
	Status  int                    `json:"status"`
	Results map[string]interface{} `json:"results"`
}

var log = logging.MustGetLogger("group_settings_cli")

const format = "%{color}[%{level}]%{color:reset} %{message}"

var backendClient *backend.Client
var routerClient *router_client.Client

var token string

func main() {
	logging.SetFormatter(logging.MustStringFormatter(format))
	logging.SetLevel(logging.ERROR, "")
	logging.SetLevel(logging.DEBUG, "group_settings_cli")

	app := cli.NewApp()
	app.Name = "group_settings_cli"
	app.Version = "0.1.0"
	app.Usage = "send commands to the lightworks_client_service"
	app.Commands = []cli.Command{
		{
			Name:      "setSettings",
			ArgsUsage: "group_id active_level vacancy_enabled inactive_level vacancy_delay_sec",
			Usage:     "set settings for a given group",
			Action:    setSettingsAction,
		},
		{
			Name:      "getSettings",
			ArgsUsage: "group_id",
			Usage:     "retrieve settings for a given group",
			Action:    getSettingsAction,
		},
		{
			Name:      "getSettingsMap",
			ArgsUsage: "group_id",
			Usage:     "retrieve entire settings map for a given site",
			Action:    getSettingsMapAction,
		},
	}

	app.Run(os.Args)
}

// Actions ---------------------------------------------------------------------

func setSettingsAction(c *cli.Context) error {
	group_id := c.Args().Get(0)
	active_level := c.Args().Get(1)
	vacancy_enabled := c.Args().Get(2)
	inactive_level := c.Args().Get(3)
	vacancy_delay_sec := c.Args().Get(4)

	request := GroupSettingsCommand{GroupID: group_id, Procedure: "setSettings",
		Params: map[string]interface{}{"active_level": active_level, "vacancy_enabled": vacancy_enabled,
			"inactive_level": inactive_level, "vacancy_delay_sec": vacancy_delay_sec}}

	log.Infof("Request: %+v", request)

	apiRequest(&request)
	// log.Infof("Status: %d", reply.Status)
	// log.Infof("Results: %+v", reply.Results)

	return nil
}

func getSettingsAction(c *cli.Context) error {
	group_id := c.Args().Get(0)

	request := GroupSettingsCommand{GroupID: group_id, Procedure: "getSettings", Params: map[string]interface{}{}}

	log.Infof("Request: %+v", request)

	apiRequest(&request)
	// log.Infof("Status: %d", reply.Status)
	// log.Infof("Results: %+v", reply.Results)

	return nil
}

func getSettingsMapAction(c *cli.Context) error {
	group_id := c.Args().Get(0)

	request := GroupSettingsCommand{GroupID: group_id, Procedure: "getSettingsMapAction", Params: map[string]interface{}{}}

	log.Infof("Request: %+v", request)

	apiRequest(&request)
	// log.Infof("Status: %d", reply.Status)
	// log.Infof("Results: %+v", reply.Results)

	return nil
}

// Helpers ---------------------------------------------------------------------

func apiRequest(command *GroupSettingsCommand) ( /*reply *ClientServiceReply, */ err error) {

	authAndConnect()

	// If we survived authAndConnect() we have a token
	command.Token = token

	uri := fmt.Sprintf("com.digitallumens.client-service.settings")
	args := structs.Map(command)

	result, err := routerClient.Call(uri, nil, args)
	if err != nil {
		log.Fatalf("routerClient.Call returned an error: %s", err.Error())
	}

	err = routerClient.Disconnect()
	if err != nil {
		log.Errorf("routerClient.Disconnect returned an error: %s", err.Error())
	}

	// GroupSettingsCommandReply
	log.Infof("Result: %+v", result)

	// reply = gw3api.NewReply()
	// err = mapstructure.Decode(result.Arguments[0], &reply)
	// if err != nil {
	//     log.Errorf("mapstructure.Decode returned an error: %s", err.Error())
	// }

	return /*reply,*/ err
}

func authAndConnect() {
	var err error
	rootPEMraw := os.Getenv("GSCLI_CLIENT_CA")
	var rootPool *x509.CertPool
	if rootPEMraw == "" {
		log.Info("No CA specified in GSCLI_CLIENT_CA. Using OS root pool.")
		rootPool = nil
	} else {
		rootPool = x509.NewCertPool()
		var rootPEM string
		if err = json.Unmarshal([]byte(rootPEMraw), &rootPEM); err != nil {
			log.Fatalf("Unmarshal rootPEM failed %s", err)
		}
		ok := rootPool.AppendCertsFromPEM([]byte(rootPEM))
		if !ok {
			log.Fatal("Unable to append CA cert to pool.")
		}
	}

	certPEMraw := os.Getenv("GSCLI_CLIENT_CERT")
	if certPEMraw == "" {
		log.Fatal("Please specify a client cert in GSCLI_CLIENT_CERT.")
	}
	var certPEM string
	if err = json.Unmarshal([]byte(certPEMraw), &certPEM); err != nil {
		log.Fatalf("Unmarshal certPEM failed %s", err)
	}

	keyPEMraw := os.Getenv("GSCLI_CLIENT_KEY")
	if keyPEMraw == "" {
		log.Fatal("Please specify a client private key in GSCLI_CLIENT_KEY")
	}
	var keyPEM string
	if err = json.Unmarshal([]byte(keyPEMraw), &keyPEM); err != nil {
		log.Fatalf("Unmarshal keyPEM failed %s", err)
	}
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		log.Fatal("Count not create X.509 key pair.")
	}

	backendClient = backend.NewClient(os.Getenv("GSCLI_BACKEND_URI"), rootPool)
	reply, err := backendClient.AuthService("group_settings_cli", cert)
	if err != nil {
		log.Fatalf("Backend auth error: %s", err)
	}
	token = reply.Token

	routerClient = router_client.NewClient(log, rootPool, cert)
	disconnected := make(chan bool)
	url := os.Getenv("GSCLI_ROUTER_URI")
	realm := "com.digitallumens"
	err = routerClient.Connect(url, realm, token, disconnected)
	if err != nil {
		log.Fatalf("Connect error: %s", err)
	}
}
