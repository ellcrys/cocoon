package client

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	context "golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"sync"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/connector/server/acl"
	"github.com/ncodes/cocoon/core/types"
	"github.com/ncodes/cstructs"
	"github.com/olekukonko/tablewriter"
	"github.com/xeonx/timeago"
)

var MaxBulkObjCount = 25

// CreateCocoon a new cocoon
func CreateCocoon(cocoon *types.Cocoon) error {

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	err = api.ValidateCocoon(cocoon)
	if err != nil {
		return err
	}

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		stopSpinner()
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	ctx := metadata.NewContext(context.Background(), metadata.Pairs("access_token", userSession.Token))
	var protoCreateCocoonReq proto.CocoonPayloadRequest
	cstructs.Copy(cocoon, &protoCreateCocoonReq)
	cocoonJSON, err := proto.NewAPIClient(conn).CreateCocoon(ctx, &protoCreateCocoonReq)
	if err != nil {
		stopSpinner()
		return err
	}

	util.FromJSON(cocoonJSON.Body, cocoon)

	stopSpinner()
	log.Info(`==> New cocoon created`)
	log.Infof(`==> Cocoon ID:  %s`, cocoon.ID)
	log.Infof(`==> Release ID: %s`, cocoon.Releases[0])

	return nil
}

// UpdateCocoon updates a cocoon and optionally creates a new
// release. A new release is created when Release fields are
// set/defined. No release is created if updated release fields match
// existing fields.
func UpdateCocoon(id string, upd *proto.CocoonPayloadRequest) error {

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	stopSpinner := util.Spinner("Please wait")

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	ctx := metadata.NewContext(context.Background(), metadata.Pairs("access_token", userSession.Token))
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	cl := proto.NewAPIClient(conn)
	resp, err := cl.UpdateCocoon(ctx, upd)
	if err != nil {
		if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
			stopSpinner()
			return fmt.Errorf("Cocoon not found: %s", id)
		}
		stopSpinner()
		return err
	}

	stopSpinner()

	var body map[string]interface{}
	if err = util.FromJSON(resp.Body, &body); err != nil {
		return fmt.Errorf("failed to coerce to map")
	}

	// no new updates was detected or performed.
	if !body["cocoonUpdated"].(bool) && len(body["newReleaseID"].(string)) == 0 {
		log.Info("No new change detected. Nothing to do")
		return nil
	}

	log.Info(`==> Update successfully applied`)

	if body["cocoonUpdated"].(bool) {
		log.Infof(`==> Cocoon has been updated`)
	}

	if len(body["newReleaseID"].(string)) != 0 {
		log.Infof(`==> New Release ID: %s`, body["newReleaseID"].(string))
	}

	return nil
}

// GetCocoons fetches one or more cocoons and logs them
func GetCocoons(ids []string) error {

	if len(ids) > MaxBulkObjCount {
		return fmt.Errorf("max number of objects exceeded. Expects a maximum of %d", MaxBulkObjCount)
	}

	var cocoons = []types.Cocoon{}
	var err error
	var resp *proto.Response
	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	for _, id := range ids {
		stopSpinner := util.Spinner("Please wait")
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		cl := proto.NewAPIClient(conn)
		resp, err = cl.GetCocoon(ctx, &proto.GetCocoonRequest{
			ID: id,
		})
		if err != nil {
			if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
				stopSpinner()
				err = fmt.Errorf("Cocoon not found: %s", id)
				break
			}
			stopSpinner()
			break
		}

		var cocoon types.Cocoon
		if err = util.FromJSON(resp.Body, &cocoon); err != nil {
			return common.JSONCoerceErr("cocoon", err)
		}

		cocoons = append(cocoons, cocoon)
		stopSpinner()
	}

	bs, _ := json.MarshalIndent(cocoons, "", "   ")
	log.Infof("%s", bs)
	if err != nil {
		return err
	}

	return nil
}

// Deploy creates and sends a deploy request to the server
func deploy(ctx context.Context, cocoonID string, useLastDeployedRelease bool) error {

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	client := proto.NewAPIClient(conn)
	resp, err := client.Deploy(ctx, &proto.DeployRequest{
		CocoonID:               cocoonID,
		UseLastDeployedRelease: useLastDeployedRelease,
	})
	if err != nil {
		return err
	} else if resp.Status != 200 {
		return fmt.Errorf("%s", resp.Body)
	}

	return nil
}

// ListCocoons fetches and displays running cocoons belonging to
// the logged in user. Set showAll to true to list both running
// and stopped cocoons.
func ListCocoons(showAll, jsonFormatted bool) error {

	var cocoons []types.Cocoon
	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()
	client := proto.NewAPIClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	resp, err := client.GetIdentity(ctx, &proto.GetIdentityRequest{
		Email: userSession.Email,
	})
	if err != nil {
		stopSpinner()
		return err
	}

	var identity types.Identity
	if err = util.FromJSON(resp.Body, &identity); err != nil {
		stopSpinner()
		return common.JSONCoerceErr("identity", err)
	}

	for _, cid := range identity.Cocoons {

		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		resp, err = client.GetCocoon(ctx, &proto.GetCocoonRequest{
			ID: cid,
		})
		if err != nil {
			stopSpinner()
			return err
		}

		var cocoon types.Cocoon
		if err = util.FromJSON(resp.Body, &cocoon); err != nil {
			stopSpinner()
			return common.JSONCoerceErr("cocoon", err)
		}

		cocoons = append(cocoons, cocoon)
	}

	stopSpinner()

	if jsonFormatted {
		bs, _ := json.MarshalIndent(cocoons, "", "   ")
		log.Info("%s", bs)
		return nil
	}

	tableData := [][]string{}
	for _, cocoon := range cocoons {
		if !showAll && !util.InStringSlice([]string{api.CocoonStatusStarted, api.CocoonStatusBuilding, api.CocoonStatusRunning}, cocoon.Status) {
			continue
		}
		created, _ := time.Parse(time.RFC3339, cocoon.CreatedAt)
		tableData = append(tableData, []string{
			cocoon.ID,
			cocoon.Releases[len(cocoon.Releases)-1],
			common.CapitalizeString(timeago.English.Format(created)),
			common.CapitalizeString(cocoon.Status),
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"COCOON ID", "RELEASE ID (LATEST)", "CREATED", "STATUS"})
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.AppendBulk(tableData)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.Render()

	return nil
}

// StopCocoon one or more running cocoon codes
func StopCocoon(ids []string) error {

	if len(ids) > MaxBulkObjCount {
		return fmt.Errorf("max number of objects exceeded. Expects a maximum of %d", MaxBulkObjCount)
	}

	var errs []error
	var stopped []string

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	cl := proto.NewAPIClient(conn)
	stopSpinner := util.Spinner("Please wait")

	for _, id := range ids {

		// find cocoon
		ctx, cc := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cc()
		resp, err := cl.GetCocoon(ctx, &proto.GetCocoonRequest{ID: id})
		if err != nil {
			if common.CompareErr(err, types.ErrCocoonNotFound) == 0 {
				errs = append(errs, fmt.Errorf("No such cocoon: %s", common.GetShortID(id)))
				continue
			}
			stopSpinner()
			return err
		}

		var cocoon types.Cocoon
		util.FromJSON(resp.GetBody(), &cocoon)
		if cocoon.Status == api.CocoonStatusStopped {
			errs = append(errs, fmt.Errorf("%s is not running", common.GetShortID(id)))
			continue
		}

		ctx, cc = context.WithTimeout(context.Background(), 1*time.Minute)
		defer cc()
		ctx = metadata.NewContext(ctx, metadata.Pairs("access_token", userSession.Token))
		_, err = cl.StopCocoon(ctx, &proto.StopCocoonRequest{ID: id})
		if err != nil {
			stopSpinner()
			return err
		}

		stopped = append(stopped, id)
	}

	stopSpinner()

	for _, id := range stopped {
		log.Info(id)
	}
	for _, err := range errs {
		log.Infof("Err: %s", err.Error())
	}

	return nil
}

// Start starts one or more new or stopped cocoon code.
// If useLastDeployedRelease is set to true, the scheduler will use the
// most recently approved and deployed release, otherwise it will
// try to deploy the latest release.
func Start(ids []string, useLastDeployedRelease bool) error {

	if len(ids) > MaxBulkObjCount {
		return fmt.Errorf("max number of objects exceeded. Expects a maximum of %d", MaxBulkObjCount)
	}

	var errs []error
	var started []string
	var muErr = sync.Mutex{}
	var muStarted = sync.Mutex{}
	var wg = sync.WaitGroup{}

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	md := metadata.Pairs("access_token", userSession.Token)
	ctx := context.Background()
	ctx = metadata.NewContext(ctx, md)

	stopSpinner := util.Spinner("Please wait")

	for _, id := range ids {
		id := id
		wg.Add(1)
		go func() {
			ctx, cc := context.WithTimeout(ctx, 1*time.Minute)
			defer cc()
			if err = deploy(ctx, id, useLastDeployedRelease); err != nil {
				muErr.Lock()
				errs = append(errs, fmt.Errorf("%s: %s", common.GetShortID(id), common.GetRPCErrDesc(err)))
				muErr.Unlock()
			} else {
				muStarted.Lock()
				started = append(started, id)
				muStarted.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()
	stopSpinner()

	if len(started) > 0 {
		log.Info("==> Successfully deployed the following:")
		for i, id := range started {
			log.Infof("==> %d. %s", i+1, id)
		}
	}

	for _, err := range errs {
		log.Infof(common.GetRPCErrDesc(err))
	}

	return nil
}

// AddSignatories adds one or more valid identities to a cocoon's signatory list.
// All valid identities are included and invalid ones will produce an error log..
func AddSignatories(cocoonID string, ids []string) error {

	if len(ids) > MaxBulkObjCount {
		return fmt.Errorf("max number of objects exceeded. Expects a maximum of %d", MaxBulkObjCount)
	}

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()

	ctx := context.Background()
	ctx = metadata.NewContext(ctx, metadata.Pairs("access_token", userSession.Token))
	cl := proto.NewAPIClient(conn)
	resp, err := cl.AddSignatories(ctx, &proto.AddSignatoriesRequest{
		CocoonID: cocoonID,
		IDs:      ids,
	})

	stopSpinner()

	if err != nil {
		return err
	}

	var r map[string][]string
	util.FromJSON(resp.Body, &r)

	if len(r["added"]) == 0 {
		log.Info("No new signatory was added")
	} else if len(r["added"]) == 1 {
		log.Info(`==> Successfully added a signatory:`)
	} else {
		log.Info(`==> Successfully added the following signatories:`)
	}

	for i, id := range r["added"] {
		log.Infof(`==> %d. %s`, i+1, id)
	}

	for _, e := range r["errs"] {
		log.Info("Err:", e)
	}

	return nil
}

// RemoveSignatories removes one or more signatories of a cocoon.
func RemoveSignatories(cocoonID string, ids []string) error {

	if len(ids) > MaxBulkObjCount {
		return fmt.Errorf("max number of objects exceeded. Expects a maximum of %d", MaxBulkObjCount)
	}

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()

	ctx := context.Background()
	ctx = metadata.NewContext(ctx, metadata.Pairs("access_token", userSession.Token))
	cl := proto.NewAPIClient(conn)
	_, err = cl.RemoveSignatories(ctx, &proto.RemoveSignatoriesRequest{
		CocoonID: cocoonID,
		IDs:      ids,
	})

	stopSpinner()

	if err != nil {
		return err
	}

	log.Info("Done.")
	return nil
}

// AddACLRule adds an acl rule to a cocoon
func AddACLRule(cocoonID, target, privileges string) error {

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	if err = common.IsValidACLTarget(target); err != nil {
		return fmt.Errorf("target format is invalid. Valid formats are: * = target any ledger a, ledgerName.[Cocoon|@Identity]")
	}

	_privileges := strings.Split(privileges, ",")
	for _, p := range _privileges {
		if !acl.IsValidPrivilege(p) {
			return fmt.Errorf("invalid privilege '%s' in '%s': ", p, privileges)
		}
	}

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()

	ctx := context.Background()
	ctx = metadata.NewContext(ctx, metadata.Pairs("access_token", userSession.Token))
	cl := proto.NewAPIClient(conn)
	_, err = cl.AddACLRule(ctx, &proto.AddACLRuleRequest{
		CocoonID:   cocoonID,
		Target:     target,
		Privileges: privileges,
	})

	if err != nil {
		stopSpinner()
		return err
	}

	stopSpinner()
	log.Info("Successfully added")

	return nil
}

// RemoveACLRule adds an acl rule to a cocoon
func RemoveACLRule(cocoonID, target string) error {

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	if err = common.IsValidACLTarget(target); err != nil {
		return fmt.Errorf("target format is invalid. Valid formats are: * = target any ledger a, ledgerName.[Cocoon|@Identity]")
	}

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()

	ctx := context.Background()
	ctx = metadata.NewContext(ctx, metadata.Pairs("access_token", userSession.Token))
	cl := proto.NewAPIClient(conn)
	_, err = cl.RemoveACLRule(ctx, &proto.RemoveACLRuleRequest{
		CocoonID: cocoonID,
		Target:   target,
	})

	if err != nil {
		stopSpinner()
		return err
	}

	stopSpinner()
	log.Info("Successfully removed")

	return nil
}

// FirewallAllow adds a firewall rule to allow connection to an outgoing destination
func FirewallAllow(dest, port, protocol string) error {
	return nil
}
