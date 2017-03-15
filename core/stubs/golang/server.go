package golang

import (
	"fmt"
	"io"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/stubs/golang/proto"
)

// StubServer defines the services of the stub's GRPC connection
type stubServer struct {
	port   int
	stream proto.Stub_TransactServer
}

// Transact listens and process invoke and response transactions from
// the connector.
func (s *stubServer) Transact(stream proto.Stub_TransactServer) error {
	s.stream = stream
	for {

		in, err := stream.Recv()
		if err == io.EOF {
			return fmt.Errorf("connection with cocoon code has ended")
		}
		if err != nil {
			return fmt.Errorf("failed to read message from connector. %s", err)
		}

		switch in.Invoke {
		case true:
			go func() {
				log.Debugf("New invoke transaction (%s) from connector", in.GetId())
				if err = s.handleInvokeTransaction(in); err != nil {
					log.Error(err.Error())
					stream.Send(&proto.Tx{
						Response: true,
						Id:       in.GetId(),
						Status:   500,
						Body:     []byte(err.Error()),
					})
				}
			}()
		case false:
			log.Debugf("New response transaction (%s) from connector", in.GetId())
			go func() {
				if err = s.handleRespTransaction(in); err != nil {
					log.Error(err.Error())
					stream.Send(&proto.Tx{
						Response: true,
						Id:       in.GetId(),
						Status:   500,
						Body:     []byte(err.Error()),
					})
				}
			}()
		}
	}
}

// handleInvokeTransaction processes invoke transaction requests
func (s *stubServer) handleInvokeTransaction(tx *proto.Tx) error {
	switch tx.GetName() {
	case "function":
		if !running {
			return fmt.Errorf("cocoon code is not running. Did you call the Run() method?")
		}

		functionName := tx.GetParams()[0]
		result, err := ccode.Invoke(tx.GetId(), functionName, tx.GetParams()[1:])
		if err != nil {
			return err
		}

		// coerce result to json
		resultJSON, err := util.ToJSON(result)
		if err != nil {
			return fmt.Errorf("failed to coerce cocoon code Invoke() result to json string. %s", err)
		}

		return s.stream.Send(&proto.Tx{
			Id:       tx.GetId(),
			Response: true,
			Status:   200,
			Body:     resultJSON,
		})

	default:
		return fmt.Errorf("Unsupported invoke transaction (%s)", tx.GetName())
	}
}

// handleRespTransaction passes the transaction to a response
// channel with a matching transaction id and deletes the channel afterwards.
func (s *stubServer) handleRespTransaction(tx *proto.Tx) error {
	if !txRespChannels.Has(tx.GetId()) {
		return fmt.Errorf("response transaction (%s) does not have a corresponding response channel", tx.GetId())
	}

	txRespCh, _ := txRespChannels.Get(tx.GetId())
	txRespCh.(chan *proto.Tx) <- tx
	txRespChannels.Remove(tx.GetId())
	return nil
}
