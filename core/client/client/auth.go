package client

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
	logging "github.com/op/go-logging"
	context "golang.org/x/net/context"
)

var log = logging.MustGetLogger("api.client")

// Login authenticates the client user. It sends the credentials
// to the platform and returns a JWT token for future requests.
func Login(email, password string) error {

	conn, err := GetAPIConnection()
	if err != nil {
		return fmt.Errorf("unable to connect to the platform")
	}
	defer conn.Close()

	stopSpinner := util.Spinner("Please wait")

	client := proto_api.NewAPIClient(conn)
	resp, err := client.Login(context.Background(), &proto_api.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		stopSpinner()
		return err
	}

	if resp.Status != 200 {
		return fmt.Errorf("%s", resp.Body)
	}

	userSession := &types.UserSession{
		Email: email,
		Token: string(resp.Body),
	}

	err = GetDefaultDB().Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("auth"))
		if err != nil {
			return err
		}
		if err = b.Put([]byte("auth.user_session"), userSession.ToJSON()); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		stopSpinner()
		return err
	}

	stopSpinner()
	log.Info("Login successful")

	return nil
}

// Logout destroy the current session. If allSessions is set,
// all sessions associated with the identity is destroyed.
func Logout(allSessions bool) error {

	currentSession, err := GetUserSessionToken()
	if err != nil {
		if common.CompareErr(err, types.ErrClientNoActiveSession) == 0 {
			return fmt.Errorf("You are not logged in")
		}
		return err
	}

	if currentSession.Token != "" || allSessions {

		stopSpinner := util.Spinner("Please wait")
		defer stopSpinner()

		conn, err := GetAPIConnection()
		if err != nil {
			return fmt.Errorf("unable to connect to the platform")
		}
		defer conn.Close()

		ctx, cc := context.WithTimeout(context.Background(), ContextTimeout)
		defer cc()
		cl := proto_api.NewAPIClient(conn)
		_, err = cl.DeleteSessions(ctx, &proto_api.DeleteSessionsRequest{
			All: allSessions,
		})
		if err != nil {
			return err
		}

		// delete current session
		if currentSession.Token != "" {
			err = GetDefaultDB().Update(func(tx *bolt.Tx) error {
				b, err := tx.CreateBucketIfNotExists([]byte("auth"))
				if err != nil {
					return err
				}
				if err = b.Delete([]byte("auth.user_session")); err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		stopSpinner()

		if allSessions {
			log.Info("Successfully logged out of all sessions")
		} else {
			log.Info("Successfully logged out of current session")
		}
	}

	return nil
}
