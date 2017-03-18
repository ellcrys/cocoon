package grpc

import (
	"fmt"
	"net"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/grpc/proto"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/orderer"
	orderer_proto "github.com/ncodes/cocoon/core/orderer/proto"
	"github.com/ncodes/cocoon/core/scheduler"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	"golang.org/x/crypto/bcrypt"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var log = logging.MustGetLogger("api.grpc")

// API defines a GRPC api for performing various
// cocoon operations such as cocoon orchestration, resource
// allocation etc
type API struct {
	server           *grpc.Server
	endedCh          chan bool
	orderDiscoTicker *time.Ticker
	orderersAddr     []string
	scheduler        scheduler.Scheduler
}

// NewAPI creates a new GRPCAPI object
func NewAPI(scheduler scheduler.Scheduler) *API {
	return &API{
		scheduler: scheduler,
	}
}

// Start starts the server
func (api *API) Start(addr string, endedCh chan bool) {

	api.endedCh = endedCh

	lis, err := net.Listen("tcp", fmt.Sprintf("%s", addr))
	if err != nil {
		log.Fatalf("failed to listen on port=%s. Err: %s", strings.Split(addr, ":")[1], err)
	}

	time.AfterFunc(2*time.Second, func() {
		log.Infof("Started server on port %s", strings.Split(addr, ":")[1])

		api.orderersAddr = orderer.DiscoverOrderers()
		if len(api.orderersAddr) > 0 {
			log.Infof("Orderer address list updated. Contains %d orderer address(es)", len(api.orderersAddr))
		} else {
			log.Warning("No orderer address was found. We won't be able to reach the orderer. ")
		}
	})

	// start a ticker to continously discover orderer addresses
	go func() {
		api.orderDiscoTicker = time.NewTicker(60 * time.Second)
		for _ = range api.orderDiscoTicker.C {
			api.orderersAddr = orderer.DiscoverOrderers()
		}
	}()

	api.server = grpc.NewServer()
	proto.RegisterAPIServer(api.server, api)
	api.server.Serve(lis)
}

// Stop stops the api and returns an exit code.
func (api *API) Stop(exitCode int) int {
	api.server.Stop()
	close(api.endedCh)
	return exitCode
}

// Deploy starts a new cocoon. The scheduler creates a job based on the requests
func (api *API) Deploy(ctx context.Context, req *proto.DeployRequest) (*proto.Response, error) {
	depInfo, err := api.scheduler.Deploy(
		req.GetId(),
		req.GetLanguage(),
		req.GetUrl(),
		req.GetReleaseTag(),
		string(req.GetBuildParam()),
		req.GetMemory(),
		req.GetCpuShare(),
	)
	if err != nil {
		if strings.HasPrefix(err.Error(), "system") {
			log.Error(err.Error())
			return nil, fmt.Errorf("failed to deploy cocoon")
		}
		return nil, err
	}

	log.Infof("Successfully deployed cocoon code %s", depInfo.ID)

	return &proto.Response{
		Id:     req.GetId(),
		Status: 200,
		Body:   []byte(depInfo.ID),
	}, nil
}

// Login authenticates a user and returns a JWT token
func (api *API) Login(ctx context.Context, req *proto.LoginRequest) (*proto.Response, error) {

	ordererConn, err := orderer.DialOrderer(api.orderersAddr)
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)
	ctx, _ = context.WithTimeout(ctx, 2*time.Minute)
	tx, err := odc.Get(ctx, &orderer_proto.GetParams{
		CocoonCodeId: "",
		Key:          fmt.Sprintf("identity.%s", req.GetEmail()),
		Ledger:       types.GetGlobalLedgerName(),
	})

	if err != nil {
		return nil, err
	} else if *tx == (orderer_proto.Transaction{}) {
		return nil, types.ErrIdentityNotFound
	}

	var value map[string]interface{}
	err = util.FromJSON([]byte(tx.GetValue()), &value)
	if err != nil {
		return nil, fmt.Errorf("failed to json encode identity data")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(value["password"].(string)), []byte(req.GetPassword())); err != nil {
		return nil, fmt.Errorf("Email or password are invalid")
	}

	claims := &jwt.MapClaims{
		"identity": tx.GetId(),
		"exp":      time.Now().AddDate(0, 1, 0).Unix(),
	}

	key := "test_key" // TODO: Important! Get this from somewhere else
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(key))
	if err != nil {
		log.Error(err.Error())
		return nil, fmt.Errorf("failed to create session token")
	}

	return &proto.Response{
		Id:     util.UUID4(),
		Status: 200,
		Body:   []byte(ss),
	}, nil
}

// CreateCocoon creates a cocoon
func (api *API) CreateCocoon(ctx context.Context, req *proto.CreateCocoonRequest) (*proto.Response, error) {

	ordererConn, err := orderer.DialOrderer(api.orderersAddr)
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)

	_, err = api.GetCocoon(ctx, &proto.GetCocoonRequest{
		Id: req.GetId(),
	})

	if err != nil && err != types.ErrCocoonNotFound {
		return nil, err
	} else if err != types.ErrCocoonNotFound {
		return nil, fmt.Errorf("cocoon with matching ID already exists")
	}

	value, _ := util.ToJSON(req)
	_, err = odc.Put(ctx, &orderer_proto.PutTransactionParams{
		CocoonCodeId: "",
		LedgerName:   types.GetGlobalLedgerName(),
		Transactions: []*orderer_proto.Transaction{
			&orderer_proto.Transaction{
				Id:        util.UUID4(),
				Key:       fmt.Sprintf("cocoon.%s", req.GetId()),
				Value:     string(value),
				CreatedAt: time.Now().Unix(),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Id:     req.GetId(),
		Status: 200,
		Body:   value,
	}, nil
}

// CreateRelease creates a release
func (api *API) CreateRelease(ctx context.Context, req *proto.CreateReleaseRequest) (*proto.Response, error) {

	ordererConn, err := orderer.DialOrderer(api.orderersAddr)
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)

	// check if release with matching ID already exists
	ctx, _ = context.WithTimeout(ctx, 2*time.Minute)
	_, err = odc.Get(ctx, &orderer_proto.GetParams{
		CocoonCodeId: "",
		Key:          fmt.Sprintf("release.%s", req.GetId()),
		Ledger:       types.GetGlobalLedgerName(),
	})

	if err != nil && common.CompareErr(err, types.ErrTxNotFound) != 0 {
		return nil, err
	} else if err != nil && common.CompareErr(err, types.ErrTxNotFound) != 0 {
		return nil, fmt.Errorf("release already exists")
	}

	value, _ := util.ToJSON(req)
	_, err = odc.Put(ctx, &orderer_proto.PutTransactionParams{
		CocoonCodeId: "",
		LedgerName:   types.GetGlobalLedgerName(),
		Transactions: []*orderer_proto.Transaction{
			&orderer_proto.Transaction{
				Id:        util.UUID4(),
				Key:       fmt.Sprintf("release.%s", req.GetId()),
				Value:     string(value),
				CreatedAt: time.Now().Unix(),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &proto.Response{
		Id:     req.GetId(),
		Status: 200,
		Body:   value,
	}, nil
}

// GetCocoon fetches a cocoon
func (api *API) GetCocoon(ctx context.Context, req *proto.GetCocoonRequest) (*proto.Response, error) {

	ordererConn, err := orderer.DialOrderer(api.orderersAddr)
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)
	tx, err := odc.Get(ctx, &orderer_proto.GetParams{
		CocoonCodeId: "",
		Key:          fmt.Sprintf("cocoon.%s", req.GetId()),
		Ledger:       types.GetGlobalLedgerName(),
	})

	if err != nil && common.CompareErr(err, types.ErrTxNotFound) != 0 {
		return nil, err
	} else if err != nil && common.CompareErr(err, types.ErrTxNotFound) == 0 {
		return nil, types.ErrCocoonNotFound
	}

	return &proto.Response{
		Id:     req.GetId(),
		Status: 200,
		Body:   []byte(tx.GetValue()),
	}, nil
}

// GetIdentity fetches an identity
func (api *API) GetIdentity(ctx context.Context, req *proto.GetIdentityRequest) (*proto.Response, error) {

	ordererConn, err := orderer.DialOrderer(api.orderersAddr)
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	odc := orderer_proto.NewOrdererClient(ordererConn)
	ctx, _ = context.WithTimeout(ctx, 2*time.Minute)
	_, err = odc.Get(ctx, &orderer_proto.GetParams{
		CocoonCodeId: "",
		Key:          fmt.Sprintf("identity.%s", req.GetEmail()),
		Ledger:       types.GetGlobalLedgerName(),
	})

	if err != nil && common.CompareErr(err, types.ErrTxNotFound) != 0 {
		return nil, err
	} else if err != nil && common.CompareErr(err, types.ErrTxNotFound) == 0 {
		return nil, types.ErrIdentityNotFound
	}

	value, _ := util.ToJSON(req)
	return &proto.Response{
		Id:     util.UUID4(),
		Status: 200,
		Body:   value,
	}, nil
}

// CreateIdentity creates a new identity
func (api *API) CreateIdentity(ctx context.Context, req *proto.CreateIdentityRequest) (*proto.Response, error) {

	ordererConn, err := orderer.DialOrderer(api.orderersAddr)
	if err != nil {
		return nil, err
	}
	defer ordererConn.Close()

	// check if identity already exists
	_, err = api.GetIdentity(ctx, &proto.GetIdentityRequest{
		Email: req.GetEmail(),
	})

	if err != nil && err != types.ErrIdentityNotFound {
		return nil, err
	} else if err == nil {
		return nil, types.ErrIdentityAlreadyExists
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	value, _ := util.ToJSON(map[string]interface{}{
		"email":    req.GetEmail(),
		"password": string(hashedPassword),
	})

	txID := util.UUID4()
	odc := orderer_proto.NewOrdererClient(ordererConn)
	ctx, _ = context.WithTimeout(ctx, 2*time.Minute)
	_, err = odc.Put(ctx, &orderer_proto.PutTransactionParams{
		CocoonCodeId: "",
		LedgerName:   types.GetGlobalLedgerName(),
		Transactions: []*orderer_proto.Transaction{
			&orderer_proto.Transaction{
				Id:        txID,
				Key:       fmt.Sprintf("identity.%s", req.GetEmail()),
				Value:     string(value),
				CreatedAt: time.Now().Unix(),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	respBody, _ := util.ToJSON(map[string]interface{}{
		"email": req.GetEmail(),
	})

	return &proto.Response{
		Id:     txID,
		Status: 200,
		Body:   respBody,
	}, nil
}
